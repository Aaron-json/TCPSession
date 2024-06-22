package controllers

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

// Member of a Session. Must be created by NewMember()
type Member struct {
	// debugging purposes. Does not have to be unique. Do not use for unique checking
	ID      int
	Session *Session
	Conn    *net.TCPConn
	Remove  func()
}

const (
	READ_BUF_SIZE     = 64 * 1024
	INTERNAL_BUF_SIZE = 5 * 1024 * 1024
)

func NewMember(ses *Session, conn *net.TCPConn, ID int) (Member, error) {
	if err := conn.SetWriteBuffer(INTERNAL_BUF_SIZE); err != nil {
		return Member{}, err
	}
	mem := Member{
		Session: ses,
		Conn:    conn,
		ID:      ID,
	}
	mem.Remove = sync.OnceFunc(
		func() {
			slog.Debug("Remove: Closing client", slog.Int("Client Id", mem.ID))
			mem.Conn.Close()
			RemoveMemberFromSession(mem)
		})

	return mem, nil
}

// Starts listening for messages on the connection and forwarding them to
// all other members in the session.
func (mem Member) Broadcast() {
	defer mem.Remove()

	buf := make([]byte, READ_BUF_SIZE)
	for {
		nr, err := mem.Conn.Read(buf)
		if err != nil {
			return
		}
		mem.Session.mu.RLock()
		for _, rec := range mem.Session.Members {
			if rec.Conn != mem.Conn {
				err = rec.Conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 200))
				if err != nil {
					slog.Debug("Broadcast: SetWriteDeadline error", slog.Any("Reason", err), slog.Int("ClientID", rec.ID))
				}
				_, err := rec.Conn.Write(buf[:nr])
				if err != nil {
					if errors.Is(err, os.ErrDeadlineExceeded) {
						// Remove tries to kick the member from the session which requires a lock.
						go rec.Remove()
						slog.Debug("Broadcast: Write Deadline error", slog.Any("Error", err), slog.Int("ClientID", rec.ID))
					} else {
						slog.Debug("Broadcast: Write error", slog.Any("Error", err), slog.Int("ClientID", rec.ID))
					}
				}

			}
		}
		mem.Session.mu.RUnlock()
	}
}
