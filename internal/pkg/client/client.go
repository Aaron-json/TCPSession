package client

import (
	"errors"
	"net"
	"sync"
)

type MessageHandler func(*Client, []byte, int)
type CloseHandler func(*Client)

// Client is a simple wrapper around a TCP connection that supports concurrent
// writes to a TCP connection.Reads have to be synchtonised by the caller.
type Client struct {
	// protect change of state. (open or closed)
	mu     sync.RWMutex
	open   bool
	conn   *net.TCPConn
	sendCh chan []byte
}

const (
	// write buffers also use this size since the buffer that is read from one client is written to other clients
	READ_BUF_SIZE = 64 * 1024
	CHAN_BUF_SIZE = 10 // TODO: test without buffering to see if buffering helps
)

// Creates a new client unopened client. To start writing messages to this client call the start method.
func NewClient(conn *net.TCPConn) *Client {
	var sendChan chan []byte
	if CHAN_BUF_SIZE <= 0 {
		sendChan = make(chan []byte)
	} else {
		sendChan = make(chan []byte, CHAN_BUF_SIZE)
	}
	c := &Client{
		conn:   conn,
		sendCh: sendChan,
		open:   false,
		// protects writing to a closed channel when user leaves
		mu: sync.RWMutex{},
	}
	return c
}

func (c *Client) Close() {
	// close connection and channel to make sure the listen and write
	// goroutines stop blocking.
	if !c.open {
		return
	}
	c.mu.Lock()
	c.open = false
	c.mu.Unlock()

	c.conn.Close()
	close(c.sendCh)

}

// Implements the Read method. After the first EOF error returned from a read call,
// the caller should close the client. The caller is responsible for synchronising
// calls to Read. It is recommended to setup a goroutine that continuously reads the connection.
func (c *Client) Read(buf []byte) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.open {
		return c.conn.Read(buf)
	} else {
		return 0, errors.New("client is not open")
	}
}
func (c *Client) writer() {
	for buf := range c.sendCh {
		c.conn.Write(buf)
	}
}

// Begins accepting reads and writes to the client. Concurrent writes are supported but concurrent reads are not.
func (c *Client) Start() {
	if c.open {
		// already opened
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	go c.writer()

	c.open = true
}

func (c *Client) Write(buf []byte) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.open {
		c.sendCh <- buf
		return len(buf), nil
	} else {
		return 0, errors.New("client is not opened")
	}
}
