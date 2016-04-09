package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/ysugimoto/aun"
	"os"
)

func main() {
	host := flag.String("h", "127.0.0.1", "Socket listen host")
	port := flag.Int("p", 12345, "Socket listen port")
	isTls := flag.Bool("tls", false, "Listen tls socket")
	pem := flag.String("pem", "", "Path to .pem file (with --tls option)")
	key := flag.String("key", "", "Path to .key file (with --tls option)")
	flag.Parse()

	server, err := aun.NewServer(*host, *port)
	if err != nil {
		fmt.Println("aun server error:", err)
		os.Exit(1)
	}

	if *isTls {
		fmt.Println("Working with TLS.")
		if _, err := os.Stat(*pem); err != nil {
			fmt.Println("pem file load error:", *pem)
			os.Exit(1)
		}
		if _, err := os.Stat(*key); err != nil {
			fmt.Println("key file load error:", *key)
			os.Exit(1)
		}
		cer, err := tls.LoadX509KeyPair(*pem, *key)
		if err != nil {
			fmt.Println("TLS error:", err)
			os.Exit(1)
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		fmt.Printf("Server listening %s:%d with TLS...\n", *host, *port)
		server.ListenTLS(1024, config)
	} else {
		fmt.Printf("Server listening %s:%d...\n", *host, *port)
		server.Listen(1024)
	}
}
