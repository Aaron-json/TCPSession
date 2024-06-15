package client

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type MessageHandler func(*Client, []byte, int)
type CloseHandler func(*Client)

// Client is a simple wrapper around a TCP connection that supports concurrent
// writes to a TCP connection.Reads have to be synchtonised by the caller.
type Client struct {
	// protect change of open state.
	mu     sync.RWMutex
	open   bool
	conn   *net.TCPConn
	sendCh chan []byte
	// keeps an estimation of how fast/slow the client is offloading data
	unread atomic.Int32
	// Called when client is closed. This function is called on a separate goroutine from the
	// client;s close method
	OnClose CloseHandler
}

const (
	// write buffers also use this size since the buffer that is read from one client is written to other clients
	CHAN_BUF_SIZE = 1 << 9 // TODO: test without buffering to see if buffering helps
	MAX_UNREAD    = 5 * 1024 * 1024
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

// Closes the client. If an OnCLose handler is set it will be called on its own goroutine.
func (c *Client) Close() {
	c.mu.Lock()
	if !c.open {
		c.mu.Unlock()
		return
	}
	c.open = false
	c.mu.Unlock()

	c.conn.Close()
	close(c.sendCh)

	if handler := c.OnClose; handler != nil {
		go handler(c)
	}
}

// Implements the Read method. After the first EOF error returned from a read call,
// the caller should close the client. The caller is responsible for synchronising
// calls to Read. It is recommended to setup a single goroutine that continuously reads the client.
func (c *Client) Read(buf []byte) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.open {
		return c.conn.Read(buf)
	} else {
		return 0, CLOSED_CLIENT
	}
}

func (c *Client) writer() {
	timer := time.NewTimer(time.Millisecond * 200)
	first := true
	done := make(chan struct{}, 1)
	for buf := range c.sendCh {
		if !first {
			timer.Reset(time.Millisecond * 200)
		}
		// time writes to the connection
		go func(c *Client, b []byte) {
			c.conn.Write(buf)
			c.unread.Add(int32(-len(buf)))
			done <- struct{}{}

		}(c, buf)

		select {
		case <-done:
			if !timer.Stop() {
				// drain the channel
				<-timer.C
			}
			log.Println("Finished write to client")
		case <-timer.C:
			log.Println("Write timed out")
			c.Close()
		}
		if first {
			first = false
		}
	}
}

// Begins accepting reads and writes to the client. Concurrent writes are supported but concurrent reads are not.
func (c *Client) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.open {
		// already opened
		return
	}
	go c.writer()
	c.open = true
}

func (c *Client) Write(buf []byte) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.open {
		// estimate unread size
		if curUnread := c.unread.Load(); curUnread >= MAX_UNREAD {
			log.Println("Slow client. Max unread size reached", curUnread)
			return 0, CLIENT_BUFFER_FULL
		}
		log.Println(len(c.sendCh))
		c.sendCh <- buf
		c.unread.Add(int32(len(buf)))
		return len(buf), nil
	} else {
		return 0, CLOSED_CLIENT
	}
}
