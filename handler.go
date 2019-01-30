package bee

// --------------------------------------------------------------------------------
type Handler interface {
	DidOpenSession(s Session)

	DidClosedSession(s Session, err error)

	DidWrittenData(s Session, data []byte)

	DidReceivedData(s Session, data []byte)
}
