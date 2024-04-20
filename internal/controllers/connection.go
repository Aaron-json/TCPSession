package controllers

import (
	"io"
	"log"
	"net"
	"time"
)

const (
	NEW_SESSION  byte = 0
	JOIN_SESSION byte = 1
)
const (
	CONN_CLOSE_DELAY time.Duration = time.Millisecond * 500
)

// wait for the first data on the connection
func HandleNewConnection(conn *net.TCPConn) {
	data, err := io.ReadAll(conn)
	if err != nil || len(data) > 1 {
		return
	}
	action := data[0]
	if action == NEW_SESSION && len(data) == 1 {
		CreateNewSession(conn)
	} else if action == JOIN_SESSION && len(data) == SESSION_CODE_LENGTH+1 {
		JoinSession(conn, string(data[1:]))
	} else {
		conn.Write([]byte{INVALID_ACTION})
		time.Sleep(CONN_CLOSE_DELAY)
		conn.Close()
	}
}
func LogDataSize(conn *net.TCPConn) error {
	buf := make([]byte, 1024*1024*10) // around 10MB
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return err
		}
		log.Println("Read size:", n, "Buffer size:", len(buf))
	}
}
