package bee

import (
	"context"
	"github.com/smartwalle/bee/conn"
	"net"
	"time"
)

// 注意: 采用了 WebSocket 标准数据包协议，只适用于使用本包内的方法创建的连接之间进行通信。
// 如果需要和使用其它语言及方式创建的连接进行通信，需要其它连接也实现标准的 WebSocket 数据包协议解析。
// --------------------------------------------------------------------------------
type Dialer struct {
	net.Dialer
	ReadBufferSize  int
	WriteBufferSize int
	WriteBufferPool conn.BufferPool
}

func (this *Dialer) Dial(network, address string) (Conn, error) {
	return this.DialContext(context.Background(), network, address)
}

func (this *Dialer) DialContext(ctx context.Context, network, address string) (Conn, error) {
	c, err := this.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	cc := NewConn(c, false, this.ReadBufferSize, this.WriteBufferSize, this.WriteBufferPool, nil, nil)
	return cc, nil
}

func Dial(network, address string) (Conn, error) {
	var d Dialer
	return d.Dial(network, address)
}

func DialTimeout(network, address string, timeout time.Duration) (Conn, error) {
	d := Dialer{}
	d.Dialer.Timeout = timeout
	return d.Dial(network, address)
}

func DialTCP(network string, laddr, raddr *net.TCPAddr) (Conn, error) {
	c, err := net.DialTCP(network, laddr, raddr)
	if err != nil {
		return nil, err
	}

	var d Dialer
	cc := NewConn(c, false, d.ReadBufferSize, d.WriteBufferSize, d.WriteBufferPool, nil, nil)
	return cc, nil
}

// --------------------------------------------------------------------------------
type Listener struct {
	net.Listener
	ReadBufferSize  int
	WriteBufferSize int
}

func (this *Listener) Accept() (Conn, error) {
	c, err := this.Listener.Accept()
	if err != nil {
		return nil, err
	}

	cc := NewConn(c, true, this.ReadBufferSize, this.WriteBufferSize, nil, nil, nil)
	return cc, nil
}

func Listen(network, address string) (*Listener, error) {
	var lc net.ListenConfig
	l, err := lc.Listen(context.Background(), network, address)
	if err != nil {
		return nil, err
	}
	return &Listener{Listener: l}, nil
}

// --------------------------------------------------------------------------------
type TCPListener struct {
	*net.TCPListener
	ReadBufferSize  int
	WriteBufferSize int
}

func (this *TCPListener) AcceptTCP() (Conn, error) {
	c, err := this.TCPListener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	cc := NewConn(c, true, this.ReadBufferSize, this.WriteBufferSize, nil, nil, nil)
	return cc, nil
}

func (this *TCPListener) Accept() (Conn, error) {
	c, err := this.TCPListener.Accept()
	if err != nil {
		return nil, err
	}
	cc := NewConn(c, true, this.ReadBufferSize, this.WriteBufferSize, nil, nil, nil)
	return cc, nil
}

func ListenTCP(network string, laddr *net.TCPAddr) (*TCPListener, error) {
	l, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, err
	}
	return &TCPListener{TCPListener: l}, nil
}
