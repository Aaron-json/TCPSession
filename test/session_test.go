package test

import (
	"errors"
	"io"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/Aaron-json/TCPSession/internal/controllers"
)

const (
	PORT = 8080
)

func TestCreatingSession(t *testing.T) {
	t.Parallel()
	conn, _, err := createSession()
	defer conn.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateAndJoinSession(t *testing.T) {
	t.Parallel()
	// create session
	conn, code, err := createSession()
	defer conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	// join the session
	conn2, err := joinSession(code)
	defer conn2.Close()
	if err != nil {
		t.Fatal(err)
	}

}
func TestSessionSendData(t *testing.T) {
	testFile, err := os.ReadFile("test_file_50mb")
	if err != nil {
		t.Fatal(err)
	}
	conns, err := createNSession(5)
	defer closeConns(conns)
	if err != nil {
		t.Fatal(err)
	}
	// for every session member, send the test file and have the other members verify the contents they receive
	for senderIdx := range len(conns) {
		n, err := conns[senderIdx].Write(testFile)
		if err != nil {
			t.Fatal("TestSessionSendData: Error wriring to connection")
		}
		if n < len(testFile) {
			t.Fatal("TestSessionSendData: Incomplete Write")
		}
		for receiverIdx, conn := range conns {
			if receiverIdx != senderIdx {
				err := ReadNAndCompare(conn, testFile)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func closeConns(conns []*net.TCPConn) {
	for _, val := range conns {
		val.Close()
	}
}

// n must be greater than 0
func createNSession(n int) ([]*net.TCPConn, error) {
	if n <= 0 {
		return nil, errors.New("createNSession: n must be greater than 0")
	}
	conns := make([]*net.TCPConn, 0, n)
	conn, code, err := createSession()
	if err != nil {
		return nil, err
	}
	conns = append(conns, conn)
	for range n - 1 { // change bounds to test if max users policy is enforced
		conn, err = joinSession(code)
		if err != nil {
			closeConns(conns)
			return nil, err
		}
		conns = append(conns, conn)
	}
	return conns, nil
}

// Helper function that reads n bytes from the TCP stream and compares it to the
// buffer contents
func ReadNAndCompare(conn *net.TCPConn, compBuf []byte) error {

	readBuf := make([]byte, len(compBuf))

	_, err := io.ReadFull(conn, readBuf)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(readBuf, compBuf) {
		return errors.New("ReadNAndCompare: File contents not equal")
	}
	return nil
}

func createConn() (*net.TCPConn, error) {
	addr := net.TCPAddr{
		IP:   net.IP{127, 0, 0, 1},
		Port: PORT,
	}
	conn, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func createSession() (*net.TCPConn, string, error) {
	conn, err := createConn()
	if err != nil {
		return nil, "", err
	}
	buf := make([]byte, controllers.SESSION_CODE_LENGTH+1)
	buf[0] = controllers.NEW_SESSION
	_, err = conn.Write(buf[0:1])
	if err != nil {
		// error writing the message to the client
		conn.Close()
		return nil, "", err
	}
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		// error reading from the connection
		conn.Close()
		return nil, "", err
	}
	if buf[0] != controllers.SUCCESS {
		conn.Close()
		return nil, "", errors.New("createSession: Server error")
	}
	return conn, string(buf[1:]), nil
}

func joinSession(code string) (*net.TCPConn, error) {
	conn, err := createConn()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, controllers.SESSION_CODE_LENGTH+1)
	buf[0] = controllers.JOIN_SESSION
	// write session code to buffer
	for i := range len(code) {
		buf[i+1] = code[i]
	}
	n, err := conn.Write(buf)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if n < len(buf) {
		conn.Close()
		return nil, errors.New("joinSession: Invalid write")
	}
	_, err = io.ReadFull(conn, buf[0:1])
	if err != nil {
		conn.Close()
		return nil, err
	}
	if buf[0] != controllers.SUCCESS {
		conn.Close()
		return nil, errors.New("joinSession: Server error")
	}
	return conn, nil
}
