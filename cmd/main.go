package main

import (
	"log"
	"net"

	"github.com/Aaron-json/TCPSession/internal/controllers"
)

func main() {
	addr := net.TCPAddr{
		Port: 8080,
		IP:   net.IP{127, 0, 0, 1},
	}
	listener, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		log.Panicln(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		go controllers.HandleNewConnection(conn)
	}
}
