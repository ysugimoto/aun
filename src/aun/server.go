package aun

import (
	"fmt"
	"net"
	"sync"
)

const TCP = "tcp"

type Server struct {
	addr        *net.TCPAddr
	socket      *net.TCPListener
	connections map[*Connection]bool
	broadcast   chan MessageReadable
	manager     chan *Connection
	join        chan *Connection
	mutex       *sync.Mutex
}

func NewServer(host string, port int) (*Server, error) {
	addr, err := net.ResolveTCPAddr(TCP, fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	return &Server{
		addr:        addr,
		connections: make(map[*Connection]bool),
		broadcast:   make(chan MessageReadable),
		manager:     make(chan *Connection),
		join:        make(chan *Connection),
		mutex:       new(sync.Mutex),
	}, nil
}

func (s *Server) Listen(maxDataSize int) (err error) {
	s.socket, err = net.ListenTCP(TCP, s.addr)
	if err != nil {
		return err
	}

	defer s.socket.Close()
	go s.acceptLoop(maxDataSize)
	for {
		select {
		case msg := <-s.broadcast:
			s.mutex.Lock()
			for c, _ := range s.connections {
				c.Write <- msg
			}
			s.mutex.Unlock()
		case c := <-s.manager:
			s.mutex.Lock()
			if _, ok := s.connections[c]; ok {
				delete(s.connections, c)
			}
			s.mutex.Unlock()
		case c := <-s.join:
			s.mutex.Lock()
			s.connections[c] = true
			s.mutex.Unlock()
		}
	}
	return nil
}

func (s *Server) acceptLoop(maxDataSize int) {
	for {
		conn, err := s.socket.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		c := NewConnection(conn, maxDataSize)
		go c.Wait(s.broadcast, s.join, s.manager)
	}
}
