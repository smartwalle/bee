package main

import "net"

func main() {
	c, err := net.Dial("tcp", ":8081")
	if err != nil {
		return
	}
	c.Write([]byte("xxx\nxxx"))
	c.Write([]byte("xxxxxx"))
	c.Write([]byte("xxxxxx"))
	c.Write([]byte("xxxxxx"))
	c.Write([]byte("\n"))

	select {}
}
