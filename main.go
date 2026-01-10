package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatalf("Error: could not open file (%v)", err)
	}
	defer file.Close()

	buffer := make([]byte, 8)

	str := ""
	for {
		n, err := file.Read(buffer)
		parts := strings.Split(string(buffer[:n]), "\n")

		for _, v := range parts[:len(parts)-1] {
			fmt.Printf("read: %s\n", str+v)
			str = ""
		}

		str += parts[len(parts)-1]

		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error: could not read %d bytes: %v", n, err)
		}
	}

	if str != "" {
		fmt.Printf("read: %s\n", str)
	}
}
