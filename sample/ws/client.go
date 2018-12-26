package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/smartwalle/bee"
)

func main() {
	var hub = bee.NewHub()
	var handler = &handler2{h: hub}

	for i := 0; i < 10000; i++ {
		c, _, _ := websocket.DefaultDialer.Dial("ws://127.0.0.1:8080/ws", nil)
		cc := bee.NewWebSocketConn(c, fmt.Sprintf("xx_%d", i), "dd", 1024, handler)
		if cc != nil {
			hub.AddConn(cc)
		}
	}

	select {}
}

type handler2 struct {
	h bee.Hub
}

func (this *handler2) DidOpenConn(c bee.Conn) {
	fmt.Println("open session", c.Identifier(), c.Tag())
}

func (this *handler2) DidClosedConn(c bee.Conn) {
	this.h.RemoveConn(c)
	fmt.Println("close session")
}

func (this *handler2) DidWrittenData(c bee.Conn, data []byte) {
	fmt.Println("write data", c.Identifier(), string(data))
}

func (this *handler2) DidReceivedData(c bee.Conn, data []byte) {
	fmt.Println("receive data", c.Identifier(), string(data))
}
