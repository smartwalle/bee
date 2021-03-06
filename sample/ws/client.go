package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/smartwalle/bee"
)

func main() {
	var hub = bee.NewHub()
	var handler = &handler2{h: hub}

	for i := 0; i < 1; i++ {
		c, _, _ := websocket.DefaultDialer.Dial("ws://127.0.0.1:8080/ws", nil)
		s := bee.NewSession(c, handler, bee.WithIdentifier(fmt.Sprintf("xx_%d", i)))
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
}

func (this *handler2) DidClosedSession(s bee.Session, err error) {
	this.h.RemoveSession(s)
	fmt.Println("close session", err)
}

func (this *handler2) DidWrittenData(s bee.Session, data []byte) {
	fmt.Println("write data", s.Identifier(), string(data))
}

func (this *handler2) DidReceivedData(s bee.Session, data []byte) {
	fmt.Println("receive data", s.Identifier(), string(data))
}
