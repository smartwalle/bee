package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/smartwalle/bee"
	"log"
	"net/http"
	"time"
)

func main() {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	var hub = bee.NewHub()
	var handler = &handler{h: hub}

	http.HandleFunc("/", home)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		var conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		var rAddr = r.RemoteAddr

		bee.NewSession(conn, rAddr, rAddr, 1024, handler)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

type handler struct {
	h bee.Hub
}

func (this *handler) DidOpenSession(s bee.Session) {
	this.h.AddSession(s)
	fmt.Println("open session", s.Identifier(), s.Tag())
	fmt.Println(this.h.Count())
}

func (this *handler) DidClosedSession(s bee.Session) {
	this.h.RemoveSession(s)
	fmt.Println("close session")
	fmt.Println(this.h.Count())
}

func (this *handler) DidWrittenData(s bee.Session, data []byte) {
	fmt.Println("write data", s.Identifier(), string(data))
}

func (this *handler) DidReceivedData(s bee.Session, data []byte) {
	fmt.Println("receive data", s.Identifier(), string(data))
	var cl = this.h.GetAllSessions()
	for _, c := range cl {
		fmt.Println(c.WriteMessage(data))
	}
	s.Write([]byte(fmt.Sprintf("%s", time.Now())))
}
