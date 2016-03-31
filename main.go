package main

import (
	"aun"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	socket, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: 10022,
	})
	if err != nil {
		log.Println(err)
		return
	}

	defer socket.Close()

	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Println(err)
			break
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	dat := make([]byte, 0)
	buf := make([]byte, 1096)
	for {
		size, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			log.Println(err)
			break
		}
		dat = append(dat, buf[:size]...)
		if size != 1096 {
			break
		}
	}

	request := aun.NewRequest(string(dat), aun.HANDSHAKE)
	if !aun.IsValidHandshake(request) {
		log.Println("Invalid")
		return
	}
	response := aun.NewResponse(request, aun.HANDSHAKE)
	conn.Write(response.ToBytes())
	conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
	conn.SetWriteDeadline(time.Now().Add(1 * time.Minute))
	for {
	}
}
