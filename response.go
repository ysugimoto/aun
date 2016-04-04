package aun

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
)

// Handshake response struct
type Response struct {
	Readable
	req *Request
}

// Create new response
func NewResponse(req *Request) *Response {
	return &Response{
		req: req,
	}
}

// Readable interface implement.
func (r *Response) getData() []byte {
	return r.Headers()
}

// Generate response headers
func (r *Response) Headers() []byte {
	buffer := []string{
		fmt.Sprintf("%s 101 Switching Protocols", r.req.Version),
		"Upgrade: websocket",
		"Connection: Upgrade",
		fmt.Sprintf("Sec-WebSocket-Accept: %s", r.genAcceptKey()),
	}

	return []byte(strings.Join(buffer, "\r\n") + "\r\n\r\n")
}

// Calcualte webosket accept key string
func (r *Response) genAcceptKey() string {
	key := strings.TrimSpace(r.req.Header("Sec-WebSocket-Key"))
	key += ACCEPTKEY
	enc := sha1.Sum([]byte(key))

	return base64.StdEncoding.EncodeToString(enc[:])
}
