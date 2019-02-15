package main

import (
	"crypto/tls"
	"fmt"
	"github.com/smartwalle/bee"
	"time"
)

func main() {
	var hub = bee.NewHub()
	var handler = &handler2{h: hub}

	for i := 0; i < 100; i++ {
		c, err := bee.DialQUIC("hk.smartwalle.tk:8889", &tls.Config{InsecureSkipVerify: true}, nil)
		if err != nil {
			fmt.Println()
			return
		}

		s := bee.NewSession(c, handler, bee.WithIdentifier(fmt.Sprintf("xx_%d", i)), bee.WithReadDeadline(time.Second*30))
		if s != nil {
			hub.AddSession(s)
		}

	}
	select {}
}

type handler2 struct {
	h bee.Hub
}

func (this *handler2) DidOpenSession(s bee.Session) {
	fmt.Println("open session", s.Identifier(), s.Tag())
	s.WriteMessage([]byte("xxx"))
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
