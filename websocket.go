package bee

import (
	"errors"
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
	if c == nil {
		return nil
	}
	var wsConn = &WebSocketConn{}
	wsConn.conn = c
	wsConn.identifier = identifier
	wsConn.tag = tag
	wsConn.maxMessageSize = maxMessageSize
	wsConn.handler = handler
	wsConn.send = make(chan []byte, 256)
	wsConn.data = make(map[string]interface{})
	wsConn.isClosed = false
	wsConn.run()
	return wsConn
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
	}()

	w.Done()

	for {
		select {
		case data, ok := <-this.send:
			this.mu.Lock()
			this.conn.SetWriteDeadline(time.Now().Add(kWriteWait))
			if !ok {
				this.conn.WriteMessage(websocket.CloseMessage, []byte{})
				this.mu.Unlock()
				return
			}

			if err := this.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				this.mu.Unlock()
				return
			}
			this.mu.Unlock()

			if this.handler != nil {
				this.handler.DidWrittenData(this, data)
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

func (this *WebSocketConn) Write(data []byte) (n int, err error) {
	this.mu.Lock()
	if this.isClosed {
		this.mu.Unlock()
		return -1, errors.New("write to closed connection")
	}

	this.conn.SetWriteDeadline(time.Now().Add(kWriteWait))

	if err = this.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		this.mu.Unlock()
		return -1, err
	}
	this.mu.Unlock()

	if this.handler != nil {
		this.handler.DidWrittenData(this, data)
	}
	return n, err
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
