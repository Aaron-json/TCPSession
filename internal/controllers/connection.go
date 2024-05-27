package controllers

import (
	"io"
	"net"
)

const (
	NEW_SESSION  byte = 0
	JOIN_SESSION byte = 1
)

// response codes
const (
	SUCCESS           byte = 0
	ERROR             byte = 1
	SESSION_NOT_FOUND byte = 2
	SESSION_FULL      byte = 3
	SERVER_FULL       byte = 4
	INVALID_ACTION    byte = 5
)

// wait for the first data on the connection
func HandleNewConnection(conn *net.TCPConn) {
	buf := make([]byte, SESSION_CODE_LENGTH+1)
	n, _ := io.ReadFull(conn, buf[0:1])
	if n == 0 {
		// read was unsuccessful
		buf[0] = ERROR
		conn.Write(buf)
		conn.Close()
		return
	}
	if action := buf[0]; action == NEW_SESSION {
		CreateNewSession(conn)
	} else if action == JOIN_SESSION {
		n, _ := io.ReadFull(conn, buf[1:])
		if n == 0 || n < SESSION_CODE_LENGTH {
			buf[0] = ERROR
			conn.Write(buf)
			conn.Close()
			return
		}
		JoinSession(conn, string(buf[1:]))
	} else {
		buf[0] = INVALID_ACTION
		conn.Write(buf)
		conn.Close()
	}
}
