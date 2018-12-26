package bee

import "net"

// --------------------------------------------------------------------------------
type Conn interface {
	Identifier() string

	Tag() string

	Set(key string, value interface{})

	Get(key string) interface{}

	Del(key string)

	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	WriteMessage(data []byte) (err error)

	Write(data []byte) (n int, err error)

	Close() error
}
