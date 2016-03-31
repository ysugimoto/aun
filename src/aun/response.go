package aun

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
)

const ACCEPTKEY = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type Response struct {
	req          *Request
	ResponseType int
}

func NewResponse(req *Request, responseType int) *Response {
	return &Response{
		req:          req,
		ResponseType: responseType,
	}
}

func (r *Response) ToBytes() []byte {
	buffer := []string{
		fmt.Sprintf("%s 101 Switching Protocols", r.req.Version),
		"Upgrade: websocket",
		"Connection: Upgrade",
		fmt.Sprintf("Sec-WebSocket-Accept: %s", r.genAcceptKey()),
	}

	fmt.Println(strings.Join(buffer, "\r\n"))
	return []byte(strings.Join(buffer, "\r\n"))
}

func (r *Response) genAcceptKey() string {
	key := strings.TrimSpace(r.req.GetHeader("Sec-WebSocket-Key"))
	key += ACCEPTKEY
	enc := sha1.Sum([]byte(key))

	return base64.StdEncoding.EncodeToString(enc[:])
}
