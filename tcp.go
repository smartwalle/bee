package bee

import (
	"net"
	"sync"
)

type TCPConn struct {
	mu         sync.Mutex
	conn       net.Conn
	identifier string
	tag        string
	handler    Handler
	send       chan []byte
	data       map[string]interface{}
	isClosed   bool
}

func NewTCPConn(c net.Conn, identifier, tag string, handler Handler) *TCPConn {
	var s = &TCPConn{}
	s.conn = c
	s.identifier = identifier
	s.tag = tag
	s.handler = handler
	s.send = make(chan []byte, 256)
	s.data = make(map[string]interface{})
	s.isClosed = false
	return s
}
