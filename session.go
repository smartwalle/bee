package bee

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type Session interface {
	Conn() *websocket.Conn
	Hub() Hub
	Identifier() string
	Write(data []byte)
	Close()
	Set(key string, value interface{})
	Get(key string) interface{}
}

type session struct {
	conn           *websocket.Conn
	identifier     string
	hub            Hub
	maxMessageSize int64
	handler        Handler
	send           chan []byte
	data           map[string]interface{}
}

func newSession(hub Hub, c *websocket.Conn, identifier string, maxMessageSize int64, handler Handler) *session {
	var s = &session{}
	s.hub = hub
	s.conn = c
	s.identifier = identifier
	s.maxMessageSize = maxMessageSize
	s.handler = handler
	s.send = make(chan []byte, 256)
	s.data = make(map[string]interface{})
	return s
}

func (this *session) read(w *sync.WaitGroup) {
	defer func() {
		close(this.send)
		this.send = nil
		this.conn.Close()
	}()

	this.conn.SetReadLimit(this.maxMessageSize)
	this.conn.SetReadDeadline(time.Now().Add(pongWait))
	this.conn.SetPongHandler(func(string) error {
		this.conn.SetReadDeadline(time.Now().Add(pongWait))
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

func (this *session) write(w *sync.WaitGroup) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		this.conn.Close()

		if this.hub != nil {
			this.hub.RemoveSession(this.identifier)
		}
		if this.handler != nil {
			this.handler.DidClosedSession(this)
		}
		this.hub = nil
		this.clean()
	}()

	w.Done()

	for {
		select {
		case msg, ok := <-this.send:
			this.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				this.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			//w, err := this.conn.NextWriter(websocket.TextMessage)
			//if err != nil {
			//	return
			//}
			//
			//w.Write(msg)
			//
			//n := len(this.send)
			//for i := 0; i < n; i++ {
			//	w.Write(newline)
			//	w.Write(<-this.send)
			//}
			//
			//if err := w.Close(); err != nil {
			//	return
			//}

			if err := this.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

			if this.handler != nil {
				this.handler.DidWrittenData(this, msg)
			}
		case <-ticker.C:
			this.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := this.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (this *session) Conn() *websocket.Conn {
	return this.conn
}

func (this *session) Hub() Hub {
	return this.hub
}

func (this *session) Identifier() string {
	return this.identifier
}

func (this *session) Write(data []byte) {
	select {
	case this.send <- data:
	default:
		if this.hub != nil {
			this.hub.RemoveSession(this.identifier)
		}
		this.hub = nil
		this.clean()
	}
}

func (this *session) Close() {
	this.conn.Close()
}

func (this *session) Set(key string, value interface{}) {
	if value != nil {
		this.data[key] = value
	}
}

func (this *session) Get(key string) interface{} {
	return this.data[key]
}

func (this *session) clean() {
	this.handler = nil
	this.data = nil
}
