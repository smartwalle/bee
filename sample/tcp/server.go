package main

import (
	"fmt"
	"github.com/smartwalle/bee"
	"time"
)

func main() {
	l, err := bee.Listen("tcp", ":8081")
	if err != nil {
		return
	}

	var hub = bee.NewHub()
	var handler = &handler{h: hub}

	for {
		c, err := l.Accept()
		if err != nil {
			return
		}

		bee.NewSession(c, c.RemoteAddr().String(), "tag", 1024, handler)
	}
}

type handler struct {
	h bee.Hub
}

func (this *handler) DidOpenSession(s bee.Session) {
	this.h.AddSession(s)
	fmt.Println("open session", s.Identifier(), s.Tag())
	fmt.Println(this.h.Len())
}

func (this *handler) DidClosedSession(s bee.Session) {
	this.h.RemoveSession(s)
	fmt.Println("close session")
	fmt.Println(this.h.Len())
}

func (this *handler) DidWrittenData(s bee.Session, data []byte) {
	fmt.Println("write data", s.Identifier(), string(data))
}

func (this *handler) DidReceivedData(s bee.Session, data []byte, err error) {
	fmt.Println("receive data", s.Identifier(), string(data))
	var cl = this.h.GetAllSessions()
	for _, c := range cl {
		fmt.Println(c.WriteMessage(data))
	}
	s.Write([]byte(fmt.Sprintf("%s", time.Now())))
}
