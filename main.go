package main

import (
	"fmt"
	"io"
	"log"
	"os"
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
				fmt.Printf("Error: could not read %d bytes: %v", n, err)
				return
			}
		}
		if str != "" {
			channel <- str
		}
	}()

	return channel
}

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatalf("Error: could not open file (%v)", err)
	}

	ch := getLinesChannel(file)
	for v := range ch {
		fmt.Printf("read: %s\n", v)
	}
}
