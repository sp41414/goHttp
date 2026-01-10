package main

import (
	"fmt"
	constants "goHttp"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	channel := make(chan string)
	buffer := make([]byte, 8)

	go func() {
		defer f.Close()
		defer close(channel)
		str := ""
		for {
			n, err := f.Read(buffer)

			if n > 0 {
				parts := strings.Split(string(buffer[:n]), "\n")
				for _, v := range parts[:len(parts)-1] {
					channel <- str + v
					str = ""
				}

				str += parts[len(parts)-1]
			}

			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Printf("Error: could not read %d bytes (%v)\n", n, err)
				return
			}
		}
		if str != "" {
			channel <- str
		}
	}()

	return channel
}

func acceptHandler(c net.Conn) {
	defer c.Close()
	defer fmt.Println("Connection closed")
	fmt.Println("Connection accepted")

	ch := getLinesChannel(c)
	for val := range ch {
		fmt.Println(val)
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
