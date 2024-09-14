package controllers

import (
	"io"
	"net"
)

// request codes
const (
	NEW_SESSION byte = iota
	JOIN_SESSION
)

// response codes
const (
	SUCCESS byte = iota
	ERROR
	SESSION_NOT_FOUND
	SESSION_FULL
	SERVER_FULL
	INVALID_ACTION
)

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
