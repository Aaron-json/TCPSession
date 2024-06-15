package controllers

import (
	"net"

	"github.com/Aaron-json/TCPSession/internal/pkg/client"
)

type Member struct {
	// Each member has a pointer to their session to avoid looking
	// it up on the global session pool and reduce its lock contention
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
	mem.Client.OnClose = func(c *client.Client) {
		RemoveMemberFromSession(mem)

	}
	return mem
}

func (sender Member) Broadcast() {
	defer func() {
		// cleanup
		sender.Client.Close()
	}()
	for {
		buf := make([]byte, READ_BUF_SIZE)
		n, err := sender.Client.Read(buf)
		if err != nil {
			break
		}
		sender.Session.mu.RLock()
		for _, mem := range sender.Session.Members {
			if mem.Client != sender.Client {
				_, err := mem.Client.Write(buf[:n])
				if err == client.CLIENT_BUFFER_FULL {
					mem.Client.Close()

				}
			}
		}
		sender.Session.mu.RUnlock()
	}
}
