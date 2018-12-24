package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/smartwalle/bee"
	"log"
	"net/http"
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

		var wsConn = bee.NewWebSocketConn(conn, rAddr, rAddr, 1024, handler)
		hub.AddConn(wsConn)
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

func (this *handler) DidOpenConn(s bee.Conn) {
	fmt.Println("open session", s.Identifier(), s.Tag())
}

func (this *handler) DidClosedConn(s bee.Conn) {
	this.h.RemoveConn(s)
	fmt.Println("close session")
}

func (this *handler) DidWrittenData(s bee.Conn, data []byte) {
	fmt.Println("write data", s.Identifier(), string(data))
}

func (this *handler) DidReceivedData(s bee.Conn, data []byte) {
	fmt.Println("receive data", s.Identifier(), string(data))
	var cl = this.h.GetAllConns()
	for _, c := range cl {
		c.Write(data)
	}
}
