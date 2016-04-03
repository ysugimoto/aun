package aun

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Check handshake request is valid
func isValidHandshake(req *Request) bool {
	if req.Method != "GET" {
		fmt.Println(1)
		return false
	}
	if req.Path == "" {
		fmt.Println(2)
		return false
	}
	if !req.HasHeader("Host") {
		fmt.Println(3)
		return false
	}
	if !req.HasHeader("Upgrade") || strings.ToLower(req.Header("Upgrade")) != "websocket" {
		fmt.Println(4)
		return false
	}
	if !req.HasHeader("Connection") || strings.ToLower(req.Header("Connection")) != "upgrade" {
		fmt.Println(5)
		return false
	}
	if !req.HasHeader("Sec-WebSocket-Key") {
		fmt.Println(6)
		return false
	}
	key, err := base64.StdEncoding.DecodeString(req.Header("Sec-WebSocket-Key"))
	if err != nil || len(key) != 16 {
		fmt.Println(7)
		return false
	}
	if !req.HasHeader("Sec-WebSocket-Version") {
		fmt.Println(8)
		return false
	}
	version, err := strconv.Atoi(req.Header("Sec-WebSocket-Version"))
	if err != nil || version != 13 {
		fmt.Println(9)
		return false
	}

	return true
}
