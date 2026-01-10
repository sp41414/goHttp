package main

import (
	"fmt"
	constants "goHttp"
	"goHttp/internal/request"
	"log"
	"net"
)

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
}

func main() {
	conn, err := net.Listen("tcp", constants.TcpPort)
	if err != nil {
		log.Fatalf("Error: tcp listener failed to start (%v)\n", err)
	}

	defer conn.Close()
	fmt.Printf("Server running on PORT %s\n", constants.TcpPort)

	for {
		accept, err := conn.Accept()
		if err != nil {
			log.Fatalf("Error: tcp listener failed to accept (%v)\n", err)
		}

		go acceptHandler(accept)
	}
}
