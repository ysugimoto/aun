package main

import (
	"aun"
	"log"
)

func main() {
	server, err := aun.NewServer("127.0.0.1", 9999)
	if err != nil {
		log.Println(err)
		return
	}
	server.Listen(1024)
}
