package aun

import (
	"errors"
	"io"
	"net"
	"time"
)

const (
	INITIALIZE = iota
	OPENING
	CONNECTED
	CLOSING
	CLOSED
)

type Connection struct {
	maxDataSize int
	state       int
	conn        net.Conn
	Read        chan *Message
	Write       chan *Message
	Close       chan struct{}
	broadcast   chan *Message
	manager     chan *Connection
	join        chan *Connection
}

func NewConnection(conn net.Conn, maxDataSize int) *Connection {
	if maxDataSize == 0 {
		maxDataSize = 1024
	}

	return &Connection{
		state:       INITIALIZE,
		maxDataSize: maxDataSize,
		conn:        conn,
		Read:        make(chan *Message),
		Write:       make(chan *Message),
		Close:       make(chan struct{}),
	}
}

func (c *Connection) Wait(broadCast chan *Message, join, manager chan *Connection) {
	c.broadcast = broadCast
	c.manager = manager
	c.join = join

	go c.readSocket()
	c.loop()
}

func (c *Connection) loop() {
	defer c.conn.Close()
OUTER:
	for {
		select {
		case msg := <-c.Read:
			switch c.state {
			case INITIALIZE:
				if err := c.handshake(msg); err == nil {
					c.join <- c
				}
			case CONNECTED:
				c.broadcast <- msg
			}
			c.conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
			c.conn.SetWriteDeadline(time.Now().Add(1 * time.Minute))
		case msg := <-c.Write:
			c.conn.Write([]byte(msg.Data))
		case <-c.Close:
			break OUTER
		}
	}
}

func (c *Connection) readSocket() {
	dat := make([]byte, 0)
	buf := make([]byte, c.maxDataSize)
	for {
		size, err := c.conn.Read(buf)
		if err != nil && err != io.EOF {
			c.Close <- struct{}{}
			return
		}
		dat = append(dat, buf[:size]...)
		if len(dat) > 0 && size != c.maxDataSize {
			c.Read <- NewMessage(dat)
		}
	}
}

func (c *Connection) handshake(msg *Message) error {
	c.state = OPENING

	request := NewRequest(string(msg.Data))
	if !isValidHandshake(request) {
		c.Close <- struct{}{}
		return errors.New("Invalid handshake request")
	}
	response := NewResponse(request)
	c.Write <- NewMessage(response.Bytes())
	return nil
}
