package aun

import (
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// Client connection socket wrapper struct.
type Connection struct {

	// connection ID (probably unique)
	Id string

	// socket max buffer size
	maxDataSize int

	// connection status
	state int

	// TCP socket connection
	conn net.Conn

	// Read message channel
	Read chan Readable

	// Send socket channel
	Write chan Readable

	// Close channel
	Close chan struct{}

	// broadcasting channnel ( supply from Server )
	broadcast chan *Frame

	// leave channnel ( supply from Server )
	manager chan *Connection

	// join channnel ( supply from Server )
	join chan *Connection

	// Frame queue stack ( treats FIN = 0 message queue )
	frameStack FrameStack
}

func generateSessionId() string {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", sha1.Sum(b))
}

// Create new connection.
func NewConnection(conn net.Conn, maxDataSize int) *Connection {
	if maxDataSize == 0 {
		// Default buffer size is 1024 bytes.
		maxDataSize = 1024
	}

	return &Connection{
		Id:          generateSessionId(),
		state:       INITIALIZE,
		maxDataSize: maxDataSize,
		conn:        conn,
		frameStack:  FrameStack{},
		Read:        make(chan Readable, 1),
		Write:       make(chan Readable, 1),
		Close:       make(chan struct{}),
	}
}

// Waiting incoming message, receive channel.
func (c *Connection) Wait(broadCast chan *Frame, join, manager chan *Connection) {
	c.broadcast = broadCast
	c.manager = manager
	c.join = join

	go c.readSocket()
	c.loop()
}

// Main channael message waiting
func (c *Connection) loop() {
	defer c.conn.Close()
	// Outer loop label
OUTER:
	for {
		select {
		// Message incoming
		case msg := <-c.Read:
			switch c.state {
			// When state is INITIALIZE, process handshake.
			case INITIALIZE:
				req := NewRequest(string(msg.getData()))
				if err := c.handshake(req, false); err == nil {
					c.join <- c
				}
			// When state is CONNECTED, incoming message.
			case CONNECTED:
				frame := NewFrame()
				err := frame.parse(msg.getData())
				if err != nil {
					fmt.Println(err)
					break OUTER
				}

				if err := c.handleFrame(frame); err != nil {
					fmt.Println(err)
					break OUTER
				}
			}
			go c.readSocket()
		// Message sending
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
		// Connection closing
		case <-c.Close:
			break OUTER
		}

		c.conn.SetDeadline(time.Now().Add(1 * time.Minute))
	}
}

// Read message from socket.
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

// Processing handshake.
func (c *Connection) handshake(request *Request, manual bool) error {
	c.state = OPENING

	// Check valid handshake request
	if !request.isValid() {
		fmt.Println(request)
		c.Close <- struct{}{}
		return errors.New("Invalid handshake request")
	}

	if !manual {
		response := NewResponse(request)
		c.Write <- response
	}
	// state changed to CONNECTED
	c.state = CONNECTED
	return nil
}

// Processing incoming message frame
func (c *Connection) handleFrame(frame *Frame) error {
	switch frame.Opcode {

	// text / binary frame
	case 1, 2:
		c.frameStack = append(c.frameStack, frame)
		if frame.Fin == 0 {
			return nil
		}
		// synthesize queueing frames (if exists)
		message := c.frameStack.synthesize()
		c.frameStack = FrameStack{}
		frames, err := BuildFrame(message, c.maxDataSize)
		if err != nil {
			return err
		}
		for _, frame := range frames {
			c.broadcast <- frame
		}

	// closing frame
	case 8:
		c.Close <- struct{}{}

	// ping frame
	case 9:
		c.broadcast <- NewPongFrame()
	}

	return nil
}
