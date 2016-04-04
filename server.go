package aun

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
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

	// Broadcast channel
	broadcast chan Readable

	// Client closing manager channel
	manager chan *Connection

	// Join the client channek
	join chan *Connection

	// Map Muten
	mutex *sync.Mutex
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
	}, nil
}

// Listen the destination host:post socket.
// First argument is max-read buffer size on messaging.
//
// Example:
//    srv := aun.NewServer("127.0.0.1", 9999)
//    src.Liten(1024) // listen with 1024 bytes message buffer
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
//    src.LitenTLS(1024, config) // listen with 1024 bytes message buffer
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

	// Loop and accepting client connection.
	// Running with goroutine
	go s.acceptLoop(maxDataSize)

	// Channel selection
	for {
		select {
		// handle the broadcast
		case msg := <-s.broadcast:
			s.mutex.Lock()
			for c, _ := range s.connections {
				c.Write <- msg
			}
			s.mutex.Unlock()

		// handle the left client
		case c := <-s.manager:
			s.mutex.Lock()
			if _, ok := s.connections[c]; ok {
				delete(s.connections, c)
			}
			s.mutex.Unlock()

		// handle the join client
		case c := <-s.join:
			s.mutex.Lock()
			s.connections[c] = true
			s.mutex.Unlock()
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
