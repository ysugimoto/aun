package aun

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	broadcast chan *Frame

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

	terminate chan os.Signal
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
		broadcast:   make(chan *Frame),
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

	s.terminate = make(chan os.Signal, 1)
	signal.Notify(s.terminate, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	s.wait()
}

func (s *Server) wait() {
MAIN:
	// Channel selection
	for {
		select {
		// handle the broadcast
		case frame := <-s.broadcast:
			s.mutex.Lock()
			if s.OnMessage != nil {
				s.OnMessage(frame.PayloadData)
			}
			for c, _ := range s.connections {
				c.Write <- NewMessage(frame.toFrameBytes())
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

		case sig := <-s.terminate:
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
		s.join <- c
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
		s.broadcast <- frame
	}

	return nil
}

// Send message to destination connection
func (s *Server) NotifyTo(message []byte, to *Connection) error {

	// Check client is connected
	if _, ok := s.connections[to]; !ok {
		return errors.New("Client not connected, abort send message.")
	}

	frames, err := BuildFrame(message, s.maxDataSize)
	if err != nil {
		return err
	}
	for _, f := range frames {
		to.Write <- NewMessage(f.toFrameBytes())
	}

	return nil
}

type HandlerServer struct {
	*Server
	callback HandlerCallback
}
type HandlerCallback func(*Connection)

func NewHandlerServer(handler HandlerCallback) *HandlerServer {
	hs := &HandlerServer{
		Server: &Server{
			connections: make(map[*Connection]bool),
			broadcast:   make(chan *Frame),
			manager:     make(chan *Connection),
			join:        make(chan *Connection),
			mutex:       new(sync.Mutex),
			Exit:        make(chan int, 2),
			maxDataSize: 4096,
		},
		callback: handler,
	}
	go hs.wait()
	return hs
}

func (hs *HandlerServer) Connect(conn net.Conn, req *Request) (*Connection, error) {
	// Create new connection, and waiting message
	c := NewConnection(conn, 4096)
	if err := c.handshake(req, true); err != nil {
		return nil, err
	}
	go c.Wait(hs.broadcast, hs.join, hs.manager)
	hs.join <- c
	return c, nil
}

func (hs *HandlerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, buf, err := w.(http.Hijacker).Hijack()
	if err != nil {
		panic(err)
	}

	// make header
	headers := make(map[string]string)
	for k, v := range r.Header {
		// http.Hijacker will change header to lowercase?
		if strings.Contains(k, "Websocket") {
			k = strings.Replace(k, "Websocket", "WebSocket", -1)
		}
		headers[k] = v[0]
	}
	// fix host header
	if _, ok := headers["Host"]; !ok {
		headers["Host"] = r.URL.Host
	}

	req := &Request{
		Method:  r.Method,
		Path:    r.URL.Path,
		Version: r.Proto,
		Headers: headers,
	}
	c, err := hs.Connect(conn, req)

	if err != nil {
		code := http.StatusForbidden
		fmt.Fprintf(buf, "HTTP/1.1 %03d %s\r\n", code, http.StatusText(code))
		buf.WriteString("\r\n")
		buf.Flush()
		conn.Close()
		return
	}

	response := NewResponse(req)
	fmt.Fprint(buf, string(response.getData()))
	buf.Flush()

	hs.callback(c)
}
