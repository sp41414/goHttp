package main

import (
	"fmt"
	"github.com/sp41414/goHttp/internal/request"
	"log"
	"net"
)

const tcpPort = ":2000"

func acceptHandler(c net.Conn) {
	defer c.Close()
	defer fmt.Println("Connection closed")
	fmt.Println("Connection accepted")

	req, err := request.RequestFromReader(c)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for k, v := range req.Headers {
		fmt.Printf("- %s: %s\n", k, v)
	}
	fmt.Printf("Body: %s\n", string(req.Body))
}

func main() {
	conn, err := net.Listen("tcp", tcpPort)
	if err != nil {
		log.Fatalf("Error: tcp listener failed to start (%v)\n", err)
	}

	defer conn.Close()
	fmt.Printf("Server running on PORT %s\n", tcpPort)

	for {
		accept, err := conn.Accept()
		if err != nil {
			log.Fatalf("Error: tcp listener failed to accept (%v)\n", err)
		}

		go acceptHandler(accept)
	}
}
