package main

import (
	"crypto/tls"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/smartwalle/bee"
	"time"
)

func main() {
	var hub = bee.NewHub()
	var handler = &handler2{h: hub}

	for i := 0; i < 1; i++ {
		c, err := bee.DialQUIC("localhost:4242", &tls.Config{InsecureSkipVerify: true}, &quic.Config{IdleTimeout: 60 * time.Second})
		if err != nil {
			return
		}

		s := bee.NewSession(c, fmt.Sprintf("xx_%d", i), "tag", 1024, handler)
		if s != nil {
			hub.AddSession(s)
		}

		s.WriteMessage([]byte("xxdd"))
	}
	select {}
}

type handler2 struct {
	h bee.Hub
}

func (this *handler2) DidOpenSession(s bee.Session) {
	fmt.Println("open session", s.Identifier(), s.Tag())
}

func (this *handler2) DidClosedSession(s bee.Session, err error) {
	this.h.RemoveSession(s)
	fmt.Println("close session")
}

func (this *handler2) DidWrittenData(s bee.Session, data []byte) {
	fmt.Println("write data", s.Identifier(), string(data))
}

func (this *handler2) DidReceivedData(s bee.Session, data []byte) {
	fmt.Println("receive data", s.Identifier(), string(data))
}
