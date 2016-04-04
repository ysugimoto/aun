package aun

import (
	"errors"
	"fmt"
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
	Read        chan MessageReadable
	Write       chan MessageReadable
	Close       chan struct{}
	broadcast   chan MessageReadable
	manager     chan *Connection
	join        chan *Connection
	frameStack  FrameStack
}

func NewConnection(conn net.Conn, maxDataSize int) *Connection {
	if maxDataSize == 0 {
		maxDataSize = 1024
	}

	return &Connection{
		state:       INITIALIZE,
		maxDataSize: maxDataSize,
		conn:        conn,
		frameStack:  FrameStack{},
		Read:        make(chan MessageReadable, 1),
		Write:       make(chan MessageReadable, 1),
		Close:       make(chan struct{}),
	}
}

func (c *Connection) Wait(broadCast chan MessageReadable, join, manager chan *Connection) {
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
				m, err := NewMessageFrame(msg.getData())
				if err != nil {
					fmt.Println(err)
					break OUTER
				}

				if err := c.handleFrame(m.Frame); err != nil {
					fmt.Println(err)
					break OUTER
				}
			}
			go c.readSocket()
		case msg := <-c.Write:
			data := msg.getData()
			size := len(data)
			var (
				written int
				err     error
			)
			for {
				if written, err = c.conn.Write(data[written:]); err != nil {
					fmt.Println(err)
					break OUTER
				}
				if written == size {
					break
				}
			}
		case <-c.Close:
			break OUTER
		}
		c.conn.SetDeadline(time.Now().Add(1 * time.Minute))
	}
	fmt.Println("connection will closing")
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
			break
		}
	}
}

func (c *Connection) handshake(msg MessageReadable) error {
	c.state = OPENING

	request := NewRequest(string(msg.getData()))
	if !isValidHandshake(request) {
		fmt.Println("Error")
		c.Close <- struct{}{}
		return errors.New("Invalid handshake request")
	}
	response := NewResponse(request)
	c.Write <- response
	c.state = CONNECTED
	return nil
}

func (c *Connection) handleFrame(frame *Frame) error {
	switch frame.Opcode {
	// text / binary frame
	case 1, 2:
		c.frameStack = append(c.frameStack, frame)
		if frame.Fin == 0 {
			return nil
		}
		message := c.frameStack.synthesize()
		c.frameStack = FrameStack{}
		frames, err := BuildFrame(message, c.maxDataSize)
		if err != nil {
			return err
		}
		for _, frame := range frames {
			c.broadcast <- NewMessage(frame.toFrameBytes())
		}
	// closing frame
	case 8:
		c.Close <- struct{}{}
	// ping frame
	case 9:
		c.broadcast <- NewMessage(NewPongFrame().toFrameBytes())
	}

	return nil
}
