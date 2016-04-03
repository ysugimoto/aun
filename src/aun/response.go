package aun

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
)

const ACCEPTKEY = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type Response struct {
	MessageReadable
	req *Request
}

func NewResponse(req *Request) *Response {
	return &Response{
		req: req,
	}
}

func (r *Response) getData() []byte {
	return r.Headers()
}

func (r *Response) Headers() []byte {
	buffer := []string{
		fmt.Sprintf("%s 101 Switching Protocols", r.req.Version),
		"Upgrade: websocket",
		"Connection: Upgrade",
		fmt.Sprintf("Sec-WebSocket-Accept: %s", r.genAcceptKey()),
	}

	return []byte(strings.Join(buffer, "\r\n") + "\r\n\r\n")
}

func (r *Response) genAcceptKey() string {
	key := strings.TrimSpace(r.req.Header("Sec-WebSocket-Key"))
	key += ACCEPTKEY
	enc := sha1.Sum([]byte(key))

	return base64.StdEncoding.EncodeToString(enc[:])
}
