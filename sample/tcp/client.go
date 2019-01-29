package main

import (
	"fmt"
	"github.com/smartwalle/bee"
)

func main() {

	var hub = bee.NewHub()
	var handler = &handler2{h: hub}

	for i := 0; i < 1000; i++ {
		c, err := bee.Dial("tcp", ":8081")
		if err != nil {
			return
		}

		s := bee.NewSession(c, fmt.Sprintf("xx_%d", i), "tag", 1024, handler)
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

func (this *handler2) DidClosedSession(s bee.Session) {
	this.h.RemoveSession(s)
	fmt.Println("close session")
}

func (this *handler2) DidWrittenData(s bee.Session, data []byte) {
	fmt.Println("write data", s.Identifier(), string(data))
}

func (this *handler2) DidReceivedData(s bee.Session, data []byte, err error) {
	fmt.Println("receive data", s.Identifier(), string(data))
}
