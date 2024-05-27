package controllers

import (
	"errors"
	"net"
	"slices"
	"sync"

	"github.com/Aaron-json/TCPSession/internal/pkg/code"
	"github.com/Aaron-json/TCPSession/internal/pkg/pool"
)

type Session struct {
	Code    string
	Members []Member
	mu      sync.RWMutex
}

const (
	MAX_SESSIONS          int = 500
	MAX_USERS_PER_SESSION int = 5
	SESSION_CODE_LENGTH   int = 7
)

var SessionPool = pool.NewPool[string, *Session](MAX_SESSIONS)

func CreateNewSession(conn *net.TCPConn) {
	newSession := NewSession("") // have not chosen code yet
	newSession.mu.Lock()
	defer newSession.mu.Unlock()
	var (
		status      byte
		sessionCode string
	)
	for {
		// find an unused session code
		sessionCode = code.Generate(SESSION_CODE_LENGTH)
		err := SessionPool.Store(sessionCode, newSession)
		if err == nil {
			newSession.Code = sessionCode
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
	buf := make([]byte, SESSION_CODE_LENGTH+1)
	buf[0] = status
	if status != SUCCESS { // could not create session
		conn.Write(buf)
		conn.Close()
	} else {
		newMember := NewMember(newSession, conn)
		newSession.Members = append(newSession.Members, newMember)
		for i := range len(sessionCode) {
			buf[i+1] = sessionCode[i]
		}
		conn.Write(buf)
		newMember.Client.Start()
		go newMember.Broadcast()
	}
}

func JoinSession(conn *net.TCPConn, sessionID string) {
	buf := make([]byte, 1)

	ses, err := SessionPool.Get(sessionID)
	if err != nil {
		if err == pool.KEY_NOT_FOUND {
			buf[0] = SESSION_NOT_FOUND
		} else {
			buf[0] = ERROR
		}
		conn.Write(buf)
		conn.Close()
		return
	}
	var (
		newMember Member
		status    byte
	)
	ses.mu.Lock()
	if len(ses.Members) == MAX_USERS_PER_SESSION {
		status = SESSION_FULL
	} else {
		newMember = NewMember(ses, conn)
		ses.Members = append(ses.Members, newMember)
		status = SUCCESS
	}
	ses.mu.Unlock()

	// handle response outside mutex. Reponse includes delay
	buf[0] = status
	if status != SUCCESS {
		conn.Write(buf)
		conn.Close()
	} else {
		conn.Write(buf)
		newMember.Client.Start()
		go newMember.Broadcast()
	}

}

func RemoveMemberFromSession(mem Member) error {
	mem.Session.mu.Lock()

	defer mem.Session.mu.Unlock()
	prevLen := len(mem.Session.Members)
	mem.Session.Members = slices.DeleteFunc(mem.Session.Members, func(v Member) bool {
		return v.Client == mem.Client
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

func NewSession(code string) *Session {
	ses := &Session{
		Code:    code,
		Members: make([]Member, 0, MAX_USERS_PER_SESSION),
		mu:      sync.RWMutex{},
	}
	return ses
}
