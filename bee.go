package bee

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

func Upgrade(upgrader websocket.Upgrader, maxMessageSize int64, hub Hub, h Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		var identifier = conn.RemoteAddr().String()
		if h != nil {
			if val := h.GetIdentifier(r); val != "" {
				identifier = val
			}
		}

		var s = newSession(hub, conn, identifier, maxMessageSize, h)

		var wg = &sync.WaitGroup{}
		wg.Add(2)

		go s.write(wg)
		go s.read(wg)

		wg.Wait()

		if h != nil {
			h.DidOpenSession(s, r)
		}
	}
}
