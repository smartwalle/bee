package bee

import (
	"net"
	"sync"
	"sync/atomic"
)

// --------------------------------------------------------------------------------
type Conn interface {
	Identifier() string

	Tag() string

	Set(key string, value interface{})

	Get(key string) interface{}

	Del(key string)

	Close() error

	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	Write(data []byte)
}

// --------------------------------------------------------------------------------
type Handler interface {
	DidOpenConn(c Conn)

	DidClosedConn(c Conn)

	DidWrittenData(c Conn, data []byte)

	DidReceivedData(c Conn, data []byte)
}

// --------------------------------------------------------------------------------
type Hub interface {
	AddConn(c Conn)

	GetConn(identifier, tag string) Conn

	GetConns(identifier string) []Conn

	GetAllConns() []Conn

	RemoveConn(c Conn)

	RemoveConns(identifier string)

	Count() int64
}

// --------------------------------------------------------------------------------
type hub struct {
	mu sync.RWMutex
	m  map[string]map[string]Conn
	c  int64
}

func NewHub() Hub {
	var h = &hub{}
	h.m = make(map[string]map[string]Conn)
	return h
}

func (this *hub) AddConn(c Conn) {
	if c != nil {
		this.mu.Lock()
		defer this.mu.Unlock()

		atomic.AddInt64(&this.c, 1)
		var cm = this.m[c.Identifier()]

		if cm == nil {
			cm = make(map[string]Conn)
			this.m[c.Identifier()] = cm
		}
		cm[c.Tag()] = c
	}
}

func (this *hub) GetConn(identifier, tag string) Conn {
	this.mu.RLock()
	defer this.mu.RUnlock()

	var cm = this.m[identifier]
	if cm != nil {
		var c = cm[tag]
		return c
	}
	return nil
}

func (this *hub) GetConns(identifier string) []Conn {
	this.mu.RLock()
	defer this.mu.RUnlock()

	var cm = this.m[identifier]
	if cm != nil {
		var cl = make([]Conn, 0, len(cm))
		for _, c := range cm {
			cl = append(cl, c)
		}
		return cl
	}
	return nil
}

func (this *hub) GetAllConns() []Conn {
	this.mu.RLock()
	defer this.mu.RUnlock()

	var cl = make([]Conn, 0, len(this.m))
	for _, cm := range this.m {
		for _, c := range cm {
			cl = append(cl, c)
		}
	}
	return cl
}

func (this *hub) RemoveConn(c Conn) {
	if c != nil {
		this.mu.Lock()
		defer this.mu.Unlock()

		var cm = this.m[c.Identifier()]
		if cm != nil {
			var c = cm[c.Tag()]
			if c != nil {
				delete(cm, c.Tag())
				atomic.AddInt64(&this.c, -1)
			}
		}
	}
}

func (this *hub) RemoveConns(identifier string) {
	this.mu.Lock()
	defer this.mu.Unlock()

	var cm = this.m[identifier]
	if cm != nil {
		delete(this.m, identifier)
		atomic.AddInt64(&this.c, -int64(len(cm)))
	}
}

func (this *hub) Count() int64 {
	return atomic.LoadInt64(&this.c)
}
