// Simple WebSocket implements over TCP/TLS.
//
// References:
//     https://tools.ietf.org/html/rfc6455
//     http://www.hcn.zaq.ne.jp/___/WEB/RFC6455-ja.html
package aun

// Connection state constants
const (
	INITIALIZE = iota
	OPENING
	CONNECTED
	CLOSING
	CLOSED
)

// Sec-WebSocket-Accept key calculate seed
const ACCEPTKEY = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
