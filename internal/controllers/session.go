package controllers

import (
	"errors"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/Aaron-json/TCPSession/internal/pkg/client"
	"github.com/Aaron-json/TCPSession/internal/pkg/code"
	"github.com/Aaron-json/TCPSession/internal/pkg/pool"
	"github.com/google/uuid"
)

const (
	SUCCESS           byte = 0
	ERROR             byte = 1
	SESSION_NOT_FOUND byte = 2
	SESSION_FULL      byte = 3
	SERVER_FULL       byte = 4
	INVALID_ACTION    byte = 5
)

type Session struct {
	Code    string
	Members []Member
	mu      sync.RWMutex
}

type Member struct {
	Id      string
	Session *Session
	Client  *client.Client
}

const (
	MAX_SESSIONS          int = 500
	MAX_USERS_PER_SESSION int = 5
	SESSION_CODE_LENGTH   int = 7
)

var SessionPool = pool.NewPool[string, *Session](MAX_SESSIONS)

func CreateNewSession(conn *net.TCPConn) {
	c, err := client.NewClient(conn)
	if err != nil {
		return
	}
	newSession := NewSession("") // have not chosen code yet
	newMember := NewMember(newSession, c)
	newSession.Members = append(newSession.Members, newMember)
	var status byte
	var sessionCode string
	for {
		// find unused code
		sessionCode = code.Generate(7)
		err = SessionPool.Store(sessionCode, newSession)
		if err == nil {
			status = SUCCESS
			break
		} else if err == pool.DUPLICATE_KEY {
			continue
		} else if err == pool.MAX_CAPACITY {
			status = SERVER_FULL
			break
		} else {
			status = ERROR
			break
		}
	}

	if status != SUCCESS {
		c.Start([]byte{status})
		// give user time to read the control message
		time.Sleep(CONN_CLOSE_DELAY)
		c.End()
	} else {
		sessionIdBytes := []byte(sessionCode)
		res := make([]byte, 0, len(sessionIdBytes)+1)
		res = append(res, status)
		res = append(res, sessionIdBytes...)
		c.Start(res)
	}
}

func JoinSession(conn *net.TCPConn, sessionID string) {
	c, err := client.NewClient(conn)
	if err != nil {
		return
	}
	var status byte
	var newMember Member
	ses, err := SessionPool.Get(sessionID)
	if err != nil {
		if err == pool.KEY_NOT_FOUND {
			status = SESSION_NOT_FOUND
		} else {
			status = ERROR
		}
	} else {
		ses.mu.Lock()
		if len(ses.Members) == MAX_USERS_PER_SESSION {
			status = SESSION_FULL
		} else {
			// no error when retrieving session
			newMember = NewMember(ses, c)
			ses.Members = append(ses.Members, newMember)
			// no error adding to pool
			status = SUCCESS
		}
		ses.mu.Unlock()
	}
	c.Start([]byte{status})
	if status != SUCCESS {
		time.Sleep(CONN_CLOSE_DELAY)
		c.End()
	}
}

func RemoveMemberFromSession(mem Member) error {
	mem.Session.mu.Lock()
	defer mem.Session.mu.Unlock()
	prevLen := len(mem.Session.Members)
	mem.Session.Members = slices.DeleteFunc(mem.Session.Members, func(v Member) bool {
		return v.Id == mem.Id
	})
	if prevLen == len(mem.Session.Members) {
		return errors.New("member not found")
	}
	if len(mem.Session.Members) == 0 {
		SessionPool.Delete(mem.Session.Code)
		return nil
	}
	return nil
}

func HandleClientClose(mem Member) error {
	err := RemoveMemberFromSession(mem)
	if err != nil {
		return err
	}
	return nil
}

func NewSession(code string) *Session {
	ses := &Session{
		Code:    code,
		Members: make([]Member, 0, MAX_USERS_PER_SESSION),
		mu:      sync.RWMutex{},
	}
	return ses
}

// Creates a new member object. Does not add the member to the session or make any modification
// to the input.
func NewMember(ses *Session, c *client.Client) Member {
	mem := Member{
		Id:      uuid.NewString(),
		Session: ses,
		Client:  c,
	}
	mem.Client.HandleMessage = func(c *client.Client, msg []byte) {
		Broadcast(mem, msg)
	}
	mem.Client.HandleClose = func(c *client.Client) {
		HandleClientClose(mem)
	}
	return mem
}
