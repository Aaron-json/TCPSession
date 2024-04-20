package client

import (
	"errors"
	"io"
	"net"
	"sync"
)

type MessageHandler func(*Client, []byte)
type CloseHandler func(*Client)

type Client struct {
	mu            sync.RWMutex
	open          bool
	conn          *net.TCPConn
	sendCh        chan []byte
	HandleClose   CloseHandler
	HandleMessage MessageHandler
}

const sendBufSize = 15

func NewClient(conn *net.TCPConn) (*Client, error) {
	c := &Client{
		conn:   conn,
		sendCh: make(chan []byte, sendBufSize),
		open:   false,
		mu:     sync.RWMutex{},
	}
	return c, nil
}

// Attempts to close the connection as soon as possible.
// Called automatically when the session is closed or when a close message is
// sent to this client. Can also be invoked manually. End must NOT be called
// more than once.
func (c *Client) End() {
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

	if c.HandleClose != nil {
		c.HandleClose(c)
	}
}

func (c *Client) listener() {
	defer c.End()
	for {

		data, err := io.ReadAll(c.conn)
		if err != nil {
			return
		}
		if c.HandleMessage != nil {
			c.HandleMessage(c, data)
		}
	}
}

func (c *Client) writer() {
	for msg := range c.sendCh {
		c.conn.Write(msg)

	}
}

// Begins receiving from and writing to the connection. Takes an optional first message parameter.
// The message is sent before the client starts listening and accepting writing to the connection.
func (c *Client) Start(msg []byte) {
	if c.open {
		// already opened
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	go c.listener()
	go c.writer()
	if msg != nil {
		c.sendCh <- msg
	}
	c.open = true
}

// This method either sends the whole message, or nothing. ie on successful write, n == len(p).
// Messeges sent before Start() or after End() methods will return an error and len == 0 and will be ignored.
func (c *Client) Send(data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.open {
		c.sendCh <- data
		return nil
	}
	return errors.New("client has been closed")
}
