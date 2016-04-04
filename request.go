package aun

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Handshale request struct.
type Request struct {
	// Request method ( always Shoud be "GET" )
	Method string

	// Request path
	Path string

	// HTTP Version ( must be greater than equal 1.1 )
	Version string

	// Request headers.
	// Parsed from socket message bytes.
	Headers map[string]string
}

// Create the new requst.
func NewRequest(message string) *Request {
	list := strings.Split(message, "\r\n")

	// first line is request line.
	// e.g. GET / HTTP/1.1
	parts := strings.Split(list[0], " ")

	// sencond-after lines are headers
	headers := make(map[string]string)
	for _, l := range list[1:] {
		if len(l) == 0 {
			continue
		}
		spl := strings.Split(l, ": ")
		headers[spl[0]] = spl[1]
	}

	return &Request{
		Method:  parts[0],
		Path:    parts[1],
		Version: parts[2],
		Headers: headers,
	}
}

// Check reuqest header has
func (r *Request) has(key string) (ok bool) {
	_, ok = r.Headers[key]

	return
}

// Get the header from key string
func (r *Request) Header(key string) (header string) {
	header, _ = r.Headers[key]

	return
}

// Check handshake request is valid
func (r *Request) isValid() bool {

	// Request method must be "GET".
	if r.Method != "GET" {
		fmt.Println(1)
		return false
	}

	// Request path must not be empty.
	if r.Path == "" {
		fmt.Println(2)
		return false
	}

	// Request must have "Host" header.
	if !r.has("Host") {
		fmt.Println(3)
		return false
	}

	// Request must have "Upgrade" header, and its value must be "websocket". (ingore case)
	if !r.has("Upgrade") || !strings.Contains(strings.ToLower(r.Header("Upgrade")), "websocket") {
		fmt.Println(4)
		return false
	}

	// Request must have "Connetion" header, and its value must contains "upgrade". (ingore case)
	if !r.has("Connection") || !strings.Contains(strings.ToLower(r.Header("Connection")), "upgrade") {
		fmt.Println(5)
		return false
	}

	// Request must have "Sec-WebSocket-Key" header.
	if !r.has("Sec-WebSocket-Key") {
		fmt.Println(6)
		return false
	}

	// "Sec-WebSocket-Key" header value must be 16 bytes length.
	key, err := base64.StdEncoding.DecodeString(r.Header("Sec-WebSocket-Key"))
	if err != nil || len(key) != 16 {
		fmt.Println(7)
		return false
	}

	// Request must have "Sec-WebSocket-Version" header.
	if !r.has("Sec-WebSocket-Version") {
		fmt.Println(8)
		return false
	}

	// "Sec-WebSocket-Version" header value must be 13.
	version, err := strconv.Atoi(r.Header("Sec-WebSocket-Version"))
	if err != nil || version != 13 {
		fmt.Println(9)
		return false
	}

	return true
}
