package aun

import (
	"io"
	"net"
)

type Connection struct {
	conn      net.Conn
	Read      chan *Message
	Write     chan *Message
	Close     chan struct{}
	broadcast chan *Message
}

func NewConnection(conn net.Conn, broadCast chan *Message) *Connection {
	return &Connection{
		conn:      conn,
		Read:      make(chan *Message),
		Write:     make(chan *Message),
		Close:     make(chan struct{}),
		broadcast: broadCast,
	}
}

func (c *Connection) Wait() {
	defer c.conn.Close()
	go c.readSocket()
	c.loop()
}

func (c *Connection) loop() {
	for {
		select {
		case msg := <-c.Read:
			c.broadcast <- msg
		case <-c.Close:
			return
		}
	}
}

func (c *Connection) readSocket() {
	dat := make([]byte, 0)
	buf := make([]byte, 2048)
	for {
		size, err := c.conn.Read(buf)
		if err != nil && err != io.EOF {
			c.Close <- struct{}{}
			return
		}
		dat = append(dat, buf[:size]...)
		if size != 2048 {
			c.Read <- NewMessage(dat)
		}
	}

}
