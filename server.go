package aun

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// TCP server with managing clients,
// boradcasting message
type Server struct {
	// TCP Address
	addr *net.TCPAddr

	// TCP Socket
	socket net.Listener

	// WebSocket clients
	connections map[*Connection]bool

	// Max buffer size per message
	maxDataSize int

	// Broadcast channel
	broadcast chan Readable

	// Client closing manager channel
	manager chan *Connection

	// Join the client channek
	join chan *Connection

	// Map Muten
	mutex *sync.Mutex

	// Exit channel
	Exit chan int

	// noop default handlers
	OnMessage MessageHandler
	OnClose   CloseHandler
	OnConnect ConnectHandler
}

// Create New WebSocket Server.
func NewServer(host string, port int) (*Server, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	return &Server{
		addr:        addr,
		connections: make(map[*Connection]bool),
		broadcast:   make(chan Readable),
		manager:     make(chan *Connection),
		join:        make(chan *Connection),
		mutex:       new(sync.Mutex),
		Exit:        make(chan int, 1),
	}, nil
}

// Listen the destination host:post socket.
// First argument is max-read buffer size on messaging.
//
// Example:
//    srv := aun.NewServer("127.0.0.1", 9999)
//    srv.Liten(1024) // listen with 1024 bytes message buffer
func (s *Server) Listen(maxDataSize int) (err error) {
	s.socket, err = net.Listen("tcp", s.addr.String())
	if err != nil {
		return err
	}

	s.serve(maxDataSize)
	return nil
}

// Listen the destination host:post socket with TLS connection.
// First argument is max-read buffer size on messaging.
//
// Example:
//    srv := aun.NewServer("127.0.0.1", 9999)
//    cer, err := tls.LoadX509KeyPair("server.pem", "server.key")
//    if err != nil {
//        log.Println(err)
//        return
//    }
//    config := &tls.Config{Certificates: []tls.Certificate{cer}}
//    srv.LitenTLS(1024, config) // listen with 1024 bytes message buffer
func (s *Server) ListenTLS(maxDataSize int, ssl *tls.Config) (err error) {
	s.socket, err = tls.Listen("tcp", s.addr.String(), ssl)
	if err != nil {
		return err
	}

	s.serve(maxDataSize)
	return nil
}

// Serve the listener.
func (s *Server) serve(maxDataSize int) {
	defer s.socket.Close()

	s.maxDataSize = maxDataSize

	// Loop and accepting client connection.
	// Running with goroutine
	go s.acceptLoop(maxDataSize)

	// observe signal event
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

MAIN:
	// Channel selection
	for {
		select {
		// handle the broadcast
		case msg := <-s.broadcast:
			s.mutex.Lock()
			if s.OnMessage != nil {
				s.OnMessage(msg.getData())
			}
			for c, _ := range s.connections {
				c.Write <- msg
			}
			s.mutex.Unlock()

		// handle the left client
		case c := <-s.manager:
			s.mutex.Lock()
			if s.OnClose != nil {
				s.OnClose(c)
			}
			if _, ok := s.connections[c]; ok {
				delete(s.connections, c)
			}
			s.mutex.Unlock()

		// handle the join client
		case c := <-s.join:
			s.mutex.Lock()
			if s.OnConnect != nil {
				s.OnConnect(c)
			}
			s.connections[c] = true
			s.mutex.Unlock()

		case <-s.Exit:
			break MAIN

		case sig := <-signalChan:
			s.handleSignal(sig)
		}
	}
}

// Thread loop accept socket connection.
func (s *Server) acceptLoop(maxDataSize int) {
	for {
		conn, err := s.socket.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		// Create new connection, and waiting message
		c := NewConnection(conn, maxDataSize)
		go c.Wait(s.broadcast, s.join, s.manager)
	}
}

// handling OS Signal
func (s *Server) handleSignal(sig os.Signal) {
	switch sig {
	case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM:
		fmt.Println("Terminating...")

		// graceful closing
		for c, _ := range s.connections {
			c.Close <- struct{}{}
		}

		s.Exit <- 1
	}
}

// Broadcast tp all clients
func (s *Server) Notify(message []byte) error {
	frames, err := BuildFrame(message, s.maxDataSize)
	if err != nil {
		return err
	}
	for _, frame := range frames {
		s.broadcast <- NewMessage(frame.toFrameBytes())
	}

	return nil
}

// Send message to destination connection
func (s *Server) NotifyTo(message []byte, to *Connection) error {

	// Check client is connected
	if _, ok := s.connections[to]; !ok {
		return errors.New("Client not connected, abort send message.")
	}

	to.Write <- NewMessage(message)

	return nil
}
