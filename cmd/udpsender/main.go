package main

import (
	"bufio"
	"goHttp/utils"
	"log"
	"net"
	"os"
)

func main() {
	raddr, err := net.ResolveUDPAddr("udp", utils.UdpPort)
	if err != nil {
		log.Fatalf("Error: could not resolve address (%v)", err)
	}

	conn, err := net.DialUDP(raddr.Network(), nil, raddr)
	if err != nil {
		log.Fatalf("Error: could not dial UDP (%v)", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		if _, err := conn.Write([]byte(line)); err != nil {
			log.Printf("write error: %v", err)
		}
	}
}
