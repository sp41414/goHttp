package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

const udpPort = "localhost:42069"

func main() {
	raddr, err := net.ResolveUDPAddr("udp", udpPort)
	if err != nil {
		log.Fatalf("Error: could not resolve address (%v)\n", err)
	}

	conn, err := net.DialUDP(raddr.Network(), nil, raddr)
	if err != nil {
		log.Fatalf("Error: could not dial UDP (%v)\n", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("read error: %v\n", err)
			continue
		}

		if _, err := conn.Write([]byte(line)); err != nil {
			log.Printf("write error: %v\n", err)
		}
	}
}
