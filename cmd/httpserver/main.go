package main

import (
	"goHttp/internal/request"
	"goHttp/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: 400,
			Message:    "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: 500,
			Message:    "Woopsie, my bad\n",
		}
	default:
		_, err := w.Write([]byte("All good, frfr\n"))
		if err != nil {
			log.Println(err)
			return &server.HandlerError{
				StatusCode: 500,
				Message:    "Woopsie, my bad\n",
			}
		}
	}
	return nil
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
