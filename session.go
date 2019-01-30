package bee

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	kWriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	kPongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than kPongWait.
	kPingPeriod = (kPongWait * 9) / 10
)

const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

//var (
//	kNewLine = []byte{'\n'}
//)

// --------------------------------------------------------------------------------
type Session interface {
	Identifier() string

	Tag() string

	Set(key string, value interface{})

	Get(key string) interface{}

	Del(key string)

	Conn() Conn

	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	WriteMessage(data []byte) (err error)

	Write(data []byte) (n int, err error)

	Close() error
}

type session struct {
	mu             sync.Mutex
	conn           Conn
	identifier     string
	tag            string
	maxMessageSize int64
	handler        Handler
	send           chan []byte
	data           map[string]interface{}
	isClosed       bool
}

func NewSession(c Conn, identifier, tag string, maxMessageSize int64, handler Handler) *session {
	if c == nil {
		return nil
	}
	var s = &session{}
	s.conn = c
	s.identifier = identifier
	s.tag = tag
	s.maxMessageSize = maxMessageSize
	s.handler = handler
	s.send = make(chan []byte, 256)
	s.data = make(map[string]interface{})
	s.isClosed = false
	s.run()
	return s
}

func (this *session) run() {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.isClosed {
		return
	}

	var w = &sync.WaitGroup{}
	w.Add(2)
	go this.write(w)
	go this.read(w)
	w.Wait()

	if this.handler != nil {
		this.handler.DidOpenSession(this)
	}
}

func (this *session) read(w *sync.WaitGroup) {
	var err error
	defer func() {
		this.close(err)
		fmt.Println("read ", err)
	}()

	this.conn.SetReadLimit(this.maxMessageSize)
	this.conn.SetReadDeadline(time.Now().Add(kPongWait))
	this.conn.SetPongHandler(func(string) error {
		this.conn.SetReadDeadline(time.Now().Add(kPongWait))
		return nil
	})

	w.Done()

	var msg []byte
	for {
		_, msg, err = this.conn.ReadMessage()

		if this.handler != nil {
			this.handler.DidReceivedData(this, msg)
		}

		if err != nil {
			return
		}
	}
}

func (this *session) write(w *sync.WaitGroup) {
	var err error
	var ticker = time.NewTicker(kPingPeriod)
	defer func() {
		ticker.Stop()
		this.close(err)
		fmt.Println("write ", err)
	}()

	w.Done()

	for {
		select {
		case data, ok := <-this.send:
			this.mu.Lock()
			if this.isClosed {
				this.mu.Unlock()
				return
			}

			this.conn.SetWriteDeadline(time.Now().Add(kWriteWait))
			if !ok {
				this.conn.WriteMessage(CloseMessage, []byte{})
				this.mu.Unlock()
				return
			}

			if err = this.conn.WriteMessage(TextMessage, data); err != nil {
				this.mu.Unlock()
				return
			}
			this.mu.Unlock()

			if this.handler != nil {
				this.handler.DidWrittenData(this, data)
			}
		case <-ticker.C:
			this.conn.SetWriteDeadline(time.Now().Add(kWriteWait))
			if err = this.conn.WriteMessage(PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (this *session) Conn() Conn {
	return this.conn
}

func (this *session) Identifier() string {
	return this.identifier
}

func (this *session) Tag() string {
	return this.tag
}

func (this *session) Set(key string, value interface{}) {
	if value != nil {
		this.data[key] = value
	}
}

func (this *session) Get(key string) interface{} {
	return this.data[key]
}

func (this *session) Del(key string) {
	delete(this.data, key)
}

func (this *session) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *session) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *session) WriteMessage(data []byte) (err error) {
	select {
	case this.send <- data:
		return nil
	default:
		err = errors.New("session is closed")
		this.close(err)
		return err
	}
}

func (this *session) Write(data []byte) (n int, err error) {
	this.mu.Lock()
	if this.isClosed {
		this.mu.Unlock()
		return -1, errors.New("session is closed")
	}

	this.conn.SetWriteDeadline(time.Now().Add(kWriteWait))

	w, err := this.conn.NextWriter(TextMessage)
	if err != nil {
		this.mu.Unlock()
		return -1, err
	}

	if n, err = w.Write(data); err != nil {
		this.mu.Unlock()
		return -1, err
	}

	if err = w.Close(); err != nil {
		this.mu.Unlock()
		return -1, err
	}

	this.mu.Unlock()

	if this.handler != nil {
		this.handler.DidWrittenData(this, data)
	}
	return n, err
}

func (this *session) Close() (err error) {
	return this.close(nil)
}

func (this *session) close(err error) (nErr error) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.isClosed {
		return nil
	}
	close(this.send)
	this.send = nil
	this.isClosed = true

	nErr = this.conn.Close()
	if this.handler != nil {
		this.handler.DidClosedSession(this, err)
	}
	this.conn = nil
	this.data = nil
	this.handler = nil
	return nErr
}
