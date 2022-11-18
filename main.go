package main

import (
	"log"
	"net"

	"github.com/didof/tcp-custom-protocol/kelly"
)

/**
# Text-based wire protocol
The data that travels on the wire is not binary but ASCII text.
The advantage of a server adopting a text-based protocol is that the client can open a TCP connection to it and interact with it by sending ASCII characters.
*/

func main() {
	ln, err := net.Listen("tcp", ":6969")
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Println("Listening on port 6969")

	hub := kelly.NewHub()
	go hub.Run()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("%v", err)
		}

		c := hub.NewClient(conn)
		go c.Read()
	}
}
