package bee

import (
	"io"
	"net"
	"time"
)

type Conn interface {
	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	Close() error

	WriteMessage(messageType int, data []byte) error

	SetReadLimit(limit int64)

	SetReadDeadline(t time.Time) error

	SetPongHandler(h func(appData string) error)

	SetWriteDeadline(t time.Time) error

	NextWriter(messageType int) (io.WriteCloser, error)

	ReadMessage() (messageType int, p []byte, err error)
}
