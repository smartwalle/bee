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

func (this *handler) DidOpenConn(c bee.Conn) {
	fmt.Println("open session", c.Identifier(), c.Tag())
	fmt.Println(this.h.Count())
}

func (this *handler) DidClosedConn(c bee.Conn) {
	this.h.RemoveConn(c)
	fmt.Println("close session")
	fmt.Println(this.h.Count())
}

func (this *handler) DidWrittenData(c bee.Conn, data []byte) {
	fmt.Println("write data", c.Identifier(), string(data))
}

func (this *handler) DidReceivedData(c bee.Conn, data []byte) {
	fmt.Println("receive data", c.Identifier(), string(data))
	var cl = this.h.GetAllConns()
	for _, c := range cl {
		c.Write(data)
	}
	c.Write([]byte(fmt.Sprintf("%s", time.Now())))
}
