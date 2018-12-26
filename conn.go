package bee

import (
	"bufio"
	"github.com/smartwalle/bee/conn"
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

func NewConn(c net.Conn, isServer bool, readBufferSize, writeBufferSize int, writeBufferPool conn.BufferPool, br *bufio.Reader, writeBuf []byte) *conn.Conn {
	return conn.NewConn(c, isServer, readBufferSize, writeBufferSize, writeBufferPool, br, writeBuf)
}
