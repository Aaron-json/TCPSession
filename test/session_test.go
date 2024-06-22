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

var (
	TEST_FILES = []string{"test_file_10kb", "test_file_10mb", "test_file_50mb"}
)

func TestCreatingSession(t *testing.T) {
	t.Parallel()
	conn, _, err := createSession()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}

func TestCreateAndJoinSession(t *testing.T) {
	t.Parallel()
	// create session
	conn, code, err := createSession()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	// join the session
	conn2, err := joinSession(code)
	defer conn2.Close()
	if err != nil {
		t.Fatal(err)
	}

}

// Tests sending and receiving data from a session with delayed reading. This test should not hang.
// and will always fail since slow readers are disconnected from the session. As long as it does not hang,
// consider it passed.
func TestSlowReader(t *testing.T) {
	testFile, err := os.ReadFile("test_file_10kb")
	if err != nil {
		t.Fatal(err)
	}
	nMembers := 5
	conns, err := createNSession(nMembers)
	if err != nil {
		t.Fatal(err)
	}
	defer closeConns(conns)
	// for every session member, send the test file and have the other members verify the contents they receive
	resCh := make(chan error, nMembers-1)
	for senderIdx := range len(conns) {
		t.Logf("TestSessionSendData: Client %v writing", senderIdx)
		n, err := conns[senderIdx].Write(testFile)
		if err != nil {
			t.Fatal("TestSessionSendData: Error wriring to connection")
		}
		if n < len(testFile) {
			t.Fatal("TestSessionSendData: Incomplete Write")
		}
		// read concurrently to avoid slow receivers
		for receiverIdx, conn := range conns {
			if receiverIdx != senderIdx {
				t.Logf("TestSessionSendData: Client %v reading", receiverIdx)
				go ReadNAndCompare(conn, testFile, resCh)
			}
		}
		// read the results from each member
		for range nMembers - 1 {
			err = <-resCh
			if err != nil {
				t.Fatal("TestSessionSendData:", err)
			}
		}
	}
}

func TestSendReceive(t *testing.T) {

	nMembers := 5
	conns, err := createNSession(nMembers)
	if err != nil {
		t.Fatal(err)
	}
	defer closeConns(conns)
	for _, filename := range TEST_FILES {
		sendReceive(conns, filename, t)
	}

}

// Function that sends a file to n members and verifies that they receive the same file.
func sendReceive(conns []*net.TCPConn, filename string, t *testing.T) {
	t.Helper()
	filedata, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	nMembers := len(conns)
	// for every session member, send the test file and have the other members verify the contents they receive
	resCh := make(chan error, nMembers-1)
	for senderIdx := range len(conns) {
		// slow receivers will get disconnected. set up readers first
		for receiverIdx, conn := range conns {
			if receiverIdx != senderIdx {
				go ReadNAndCompare(conn, filedata, resCh)
				t.Logf("TestSendReceive: Client %v reading", receiverIdx)
			}
		}
		t.Logf("TestSendReceive: Client %v writing", senderIdx)
		n, err := conns[senderIdx].Write(filedata)
		if err != nil {
			t.Fatal("TestSendReceive: Error wriring to connection")
		}
		if n < len(filedata) {
			t.Fatal("TestSendReceive: Incomplete Write")
		}

		// read the results from each member
		for range nMembers - 1 {
			err = <-resCh
			if err != nil {
				t.Fatal("TestSendReceive:", err)
			}
		}
	}
}
func closeConns(conns []*net.TCPConn) {
	for _, val := range conns {
		val.Close()
	}
}

// Creates a session of n members. If any of them fails, it will close all connections and return an error.
// n must be greater than 0.
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
func ReadNAndCompare(conn *net.TCPConn, compBuf []byte, resCh chan error) {

	readBuf := make([]byte, len(compBuf))

	_, err := io.ReadFull(conn, readBuf)
	if err != nil {
		resCh <- err
	} else if !reflect.DeepEqual(readBuf, compBuf) {
		resCh <- errors.New("ReadNAndCompare: File contents not equal")
	} else {
		resCh <- nil
	}
}

func createConn() (*net.TCPConn, error) {
	addr := net.TCPAddr{
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
