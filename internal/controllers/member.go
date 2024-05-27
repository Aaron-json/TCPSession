package controllers

import (
	"net"

	"github.com/Aaron-json/TCPSession/internal/pkg/client"
)

type Member struct {
	Session *Session
	Client  *client.Client
}

const (
	READ_BUF_SIZE = 64 * 1024
)

func NewMember(ses *Session, conn *net.TCPConn) Member {
	mem := Member{
		Session: ses,
		Client:  client.NewClient(conn),
	}

	return mem
}

func (sender Member) Broadcast() {
	defer func() {
		// cleanup
		sender.Client.Close()
		RemoveMemberFromSession(sender)
	}()
	for {
		buf := make([]byte, READ_BUF_SIZE)
		n, err := sender.Client.Read(buf)
		if err != nil {
			break
		}
		sender.Session.mu.RLock()
		counter := 0
		for _, mem := range sender.Session.Members {
			if mem.Client != sender.Client {
				counter++
				mem.Client.Write(buf[:n])
			}
		}
		sender.Session.mu.RUnlock()
	}
}
