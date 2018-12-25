package bee

// --------------------------------------------------------------------------------
type Handler interface {
	DidOpenConn(c Conn)

	DidClosedConn(c Conn)

	DidWrittenData(c Conn, data []byte)

	DidReceivedData(c Conn, data []byte)
}
