package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		return
	}

	for {
		c, err := l.Accept()
		if err != nil {
			return
		}
		fmt.Println(c)

		var r = bufio.NewReader(c)
		for {

			fmt.Println(r.ReadBytes(byte('\n')))
		}
	}
}
