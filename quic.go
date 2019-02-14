package bee

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/lucas-clemente/quic-go"
	"github.com/smartwalle/bee/conn"
	"net"
)

// --------------------------------------------------------------------------------
type QUICDialer struct {
	ReadBufferSize  int
	WriteBufferSize int
	WriteBufferPool conn.BufferPool
	tlsConf         *tls.Config
	config          *quic.Config
}

func NewQUICDialer(tlsConf *tls.Config, config *quic.Config) *QUICDialer {
	var d = &QUICDialer{}
	d.tlsConf = tlsConf
	d.config = config
	return d
}

func (this *QUICDialer) Dial(network, addr string) (Conn, error) {
	return this.DialContext(context.Background(), network, addr)
}

func (this *QUICDialer) DialContext(ctx context.Context, network, addr string) (Conn, error) {
	sess, err := quic.DialAddrContext(ctx, addr, this.tlsConf, this.config)
	if err != nil {
		return nil, err
	}

	stream, err := sess.OpenStream()
	if err != nil {
		sess.Close()
		return nil, err
	}

	if stream == nil {
		sess.Close()
		return nil, errors.New("closed stream")
	}

	c := &qSession{sess: sess, Stream: stream}
	return NewConn(c, false, this.ReadBufferSize, this.WriteBufferSize, nil, nil, nil), nil
}

func DialQUIC(addr string, tlsConf *tls.Config, config *quic.Config) (Conn, error) {
	var d QUICDialer
	d.tlsConf = tlsConf
	d.config = config
	return d.DialContext(context.Background(), "", addr)
}

// --------------------------------------------------------------------------------
type QUICListener struct {
	ln              quic.Listener
	acceptConn      chan *qConn
	ReadBufferSize  int
	WriteBufferSize int
}

func (this *QUICListener) doAccept() {
	for {
		sess, err := this.ln.Accept()
		if err != nil {
			return
		}

		go func(sess quic.Session) {
			for {
				stream, err := sess.AcceptStream()
				if err != nil {
					sess.Close()
					return
				}

				if stream == nil {
					sess.Close()
					return
				}

				this.acceptConn <- &qConn{
					conn: &qSession{sess: sess, Stream: stream},
					err:  nil,
				}
			}
		}(sess)
	}
}

func (this *QUICListener) Accept() (Conn, error) {
	ac := <-this.acceptConn
	if ac.err != nil {
		return nil, ac.err
	}
	return NewConn(ac.conn, true, this.ReadBufferSize, this.WriteBufferSize, nil, nil, nil), nil
}

func ListenQUIC(addr string, tlsConf *tls.Config, config *quic.Config) (*QUICListener, error) {
	l, err := quic.ListenAddr(addr, tlsConf, config)
	if err != nil {
		return nil, err
	}

	ln := &QUICListener{ln: l, acceptConn: make(chan *qConn, 1)}
	go ln.doAccept()
	return ln, nil
}

// --------------------------------------------------------------------------------
type qConn struct {
	conn net.Conn
	err  error
}

// --------------------------------------------------------------------------------
type qSession struct {
	sess quic.Session
	quic.Stream
}

func (this *qSession) LocalAddr() net.Addr {
	return this.sess.LocalAddr()
}

func (this *qSession) RemoteAddr() net.Addr {
	return this.sess.RemoteAddr()
}

func (this *qSession) Close() error {
	this.Stream.Close()
	return this.sess.Close()
}
