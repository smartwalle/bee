package bee

import "net/http"

type Handler interface {
	GetIdentifier(r *http.Request) string

	DidOpenSession(s Session, r *http.Request)

	DidClosedSession(s Session)

	DidWrittenData(s Session, data []byte)

	DidReceivedData(s Session, data []byte)
}
