package main

import (
	"log"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/Aaron-json/TCPSession/internal/controllers"
)

func main() {
	// set default logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)

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
