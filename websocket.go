package bee

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
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

type WebSocketConn struct {
	mu             sync.Mutex
	conn           *websocket.Conn
	identifier     string
	tag            string
	maxMessageSize int64
	handler        Handler
	send           chan []byte
	data           map[string]interface{}
	isClosed       bool
}

func NewWebSocketConn(c *websocket.Conn, identifier, tag string, maxMessageSize int64, handler Handler) *WebSocketConn {
	var s = &WebSocketConn{}
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

func (this *WebSocketConn) run() {
	var wg = &sync.WaitGroup{}
	wg.Add(2)

	go this.write(wg)
	go this.read(wg)

	wg.Wait()

	if this.handler != nil {
		this.handler.DidOpenConn(this)
	}
}

func (this *WebSocketConn) read(w *sync.WaitGroup) {
	defer func() {
		this.Close()

		fmt.Println("read defer")
	}()

	this.conn.SetReadLimit(this.maxMessageSize)
	this.conn.SetReadDeadline(time.Now().Add(kPongWait))
	this.conn.SetPongHandler(func(string) error {
		this.conn.SetReadDeadline(time.Now().Add(kPongWait))
		return nil
	})

	w.Done()

	for {
		_, msg, err := this.conn.ReadMessage()

		if err != nil {
			break
		}

		if this.handler != nil {
			this.handler.DidReceivedData(this, msg)
		}
	}
}

func (this *WebSocketConn) write(w *sync.WaitGroup) {
	ticker := time.NewTicker(kPingPeriod)
	defer func() {
		ticker.Stop()
		this.Close()

		fmt.Println("write defer")
	}()

	w.Done()

	for {
		select {
		case msg, ok := <-this.send:
			this.conn.SetWriteDeadline(time.Now().Add(kWriteWait))
			if !ok {
				this.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := this.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

			if this.handler != nil {
				this.handler.DidWrittenData(this, msg)
			}
		case <-ticker.C:
			this.conn.SetWriteDeadline(time.Now().Add(kWriteWait))
			if err := this.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (this *WebSocketConn) Conn() *websocket.Conn {
	return this.conn
}

func (this *WebSocketConn) Close() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.isClosed {
		return nil
	}
	close(this.send)
	this.send = nil
	if this.handler != nil {
		this.handler.DidClosedConn(this)
	}
	this.handler = nil
	this.data = nil
	this.isClosed = true
	return this.conn.Close()
}

func (this *WebSocketConn) Identifier() string {
	return this.identifier
}

func (this *WebSocketConn) Tag() string {
	return this.tag
}

func (this *WebSocketConn) Set(key string, value interface{}) {
	if value != nil {
		this.data[key] = value
	}
}

func (this *WebSocketConn) Get(key string) interface{} {
	return this.data[key]
}

func (this *WebSocketConn) Del(key string) {
	delete(this.data, key)
}

func (this *WebSocketConn) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *WebSocketConn) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *WebSocketConn) WriteMessage(data []byte) (err error) {
	select {
	case this.send <- data:
		return nil
	default:
		this.Close()
		return errors.New("write to closed connection")
	}
}
