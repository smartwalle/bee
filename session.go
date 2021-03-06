package bee

import (
	"errors"
	"net"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	kDefaultWriteDeadline = 10 * time.Second

	kDefaultReadDeadline = 60 * time.Second

	// Time allowed to read the next pong message from the peer.
	//kDefaultPongWait = kDefaultReadDeadline

	// Send pings to peer with this period. Must be less than kDefaultPongWait.
	//kDefaultPingPeriod = (kDefaultPongWait * 9) / 10

	kDefaultWriteBufferSize = 8

	kDefaultMaxMessageSize = 1024

	kDefaultTag = "default"
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
type Option interface {
	Apply(*session)
}

type optionFunc func(*session)

func (f optionFunc) Apply(s *session) {
	f(s)
}

func WithWriteBufferSize(size int) Option {
	return optionFunc(func(s *session) {
		if size <= 0 {
			s.writeBufferSize = kDefaultWriteBufferSize
		}
		s.writeBufferSize = size
	})
}

func WithMaxMessageSize(size int64) Option {
	return optionFunc(func(s *session) {
		if size <= 0 {
			size = kDefaultMaxMessageSize
		}
		s.maxMessageSize = size
	})
}

func WithIdentifier(identifier string) Option {
	return optionFunc(func(s *session) {
		if identifier == "" {
			identifier = s.conn.RemoteAddr().String()
		}
		s.identifier = identifier
	})
}

func WithTag(tag string) Option {
	return optionFunc(func(s *session) {
		if tag == "" {
			tag = kDefaultTag
		}
		s.tag = tag
	})
}

func WithWriteDeadline(t time.Duration) Option {
	return optionFunc(func(s *session) {
		if t <= 0 {
			t = kDefaultWriteDeadline
		}
		s.writeDeadline = t
	})
}

func WithReadDeadline(t time.Duration) Option {
	return optionFunc(func(s *session) {
		if t <= 0 {
			t = kDefaultReadDeadline
		}
		s.readDeadline = t
	})
}

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
	mu      sync.Mutex
	conn    Conn
	handler Handler

	identifier      string
	tag             string
	maxMessageSize  int64
	writeBufferSize int

	writeDeadline time.Duration
	readDeadline  time.Duration

	pongWait   time.Duration
	pingPeriod time.Duration

	send     chan []byte
	data     map[string]interface{}
	isClosed bool
}

func NewSession(c Conn, handler Handler, opts ...Option) *session {
	if c == nil {
		return nil
	}
	var s = &session{}
	s.conn = c
	s.handler = handler
	s.identifier = s.conn.RemoteAddr().String()
	s.tag = kDefaultTag
	s.maxMessageSize = kDefaultMaxMessageSize
	s.writeBufferSize = kDefaultWriteBufferSize

	s.writeDeadline = kDefaultWriteDeadline
	s.readDeadline = kDefaultReadDeadline

	for _, opt := range opts {
		opt.Apply(s)
	}

	s.pongWait = s.readDeadline
	s.pingPeriod = (s.pongWait * 9) / 10

	s.send = make(chan []byte, s.writeBufferSize)
	s.data = make(map[string]interface{})
	s.isClosed = false
	s.run()
	return s
}

func (this *session) run() {
	this.mu.Lock()

	if this.isClosed {
		this.mu.Unlock()
		return
	}

	var w = &sync.WaitGroup{}
	w.Add(2)
	go this.write(w)
	go this.read(w)
	w.Wait()
	this.mu.Unlock()

	if this.handler != nil {
		this.handler.DidOpenSession(this)
	}
}

func (this *session) read(w *sync.WaitGroup) {
	var err error
	defer func() {
		this.close(err)
	}()

	this.conn.SetReadLimit(this.maxMessageSize)
	this.conn.SetReadDeadline(time.Now().Add(this.pongWait))
	this.conn.SetPongHandler(func(string) error {
		this.conn.SetReadDeadline(time.Now().Add(this.pongWait))
		return nil
	})

	w.Done()

	var msg []byte
	for {
		if this.isClosed {
			return
		}
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
	var ticker = time.NewTicker(this.pingPeriod)
	defer func() {
		ticker.Stop()
		this.close(err)
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

			this.conn.SetWriteDeadline(time.Now().Add(this.writeDeadline))
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
			this.mu.Lock()
			if this.isClosed {
				this.mu.Unlock()
				return
			}
			this.mu.Unlock()

			this.conn.SetWriteDeadline(time.Now().Add(this.writeDeadline))
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

	this.conn.SetWriteDeadline(time.Now().Add(this.writeDeadline))

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
