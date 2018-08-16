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

	http.HandleFunc("/", home)
	http.HandleFunc("/ws", bee.Upgrade(upgrader, 1024, bee.NewHub(), &handler{}))
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
}

func (this *handler) GetIdentifier(r *http.Request) string {
	return ""
}

func (this *handler) DidOpenSession(s bee.Session, r *http.Request) {
	fmt.Println("open session", s.Identifier())
}

func (this *handler) DidClosedSession(s bee.Session) {
	fmt.Println("close session")
}

func (this *handler) DidWrittenData(s bee.Session, data []byte) {
	fmt.Println("write data", s.Identifier(), string(data))
}

func (this *handler) DidReceivedData(s bee.Session, data []byte) {
	fmt.Println("receive data", s.Identifier(), string(data))

	s.Hub().Range(func(identifier string, s bee.Session) bool {
		s.Write(data)
		return true
	})
}
