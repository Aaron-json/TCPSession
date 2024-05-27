package main

import (
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/Aaron-json/TCPSession/internal/controllers"
)

func main() {
	port := 8080
	addr := net.TCPAddr{
		Port: port,
		IP:   net.IP{127, 0, 0, 1},
	}
	//set up server for profiling
	go pprofServer()
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

func pprofServer() {
	http.ListenAndServe("127.0.0.1:8081", nil)
}
