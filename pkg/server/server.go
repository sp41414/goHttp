// Package server implements a high-performance TCP-based HTTP/1.1 server.
// It manages the connection lifecycle, concurrent request handling,
// and provides a simple interface for custom request handlers.
package server

import (
	"fmt"
	"github.com/sp41414/goHttp/pkg/request"
	"github.com/sp41414/goHttp/pkg/response"
	"log"
	"net"
	"sync/atomic"
)

// Handler is a function type that processes an incoming HTTP request
// and writes the response back to the client.
//
// Every handler is executed in its own goroutine, allowing the server
// to process multiple connections concurrently.
type Handler func(w *response.Writer, req *request.Request)

// HandlerError represents an application-level error occurring during
// request processing, associated with an HTTP status code.
type HandlerError struct {
	StatusCode int
	Message    string
}

// Server represents an active HTTP server instance listening for connections.
type Server struct {
	Listener net.Listener
	Closed   atomic.Bool // Closed tracks the server's shutdown status safely.
}

// Serve initializes and starts a new HTTP server on the specified port.
// It returns a pointer to the Server instance and begins listening
// for connections in a background goroutine.
//
// Example:
//
//	s, err := server.Serve(8080, handler)
func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		Listener: listener,
	}
	go s.listen(handler)
	return s, nil
}

// Close gracefully stops the server by closing the underlying TCP listener.
// Any ongoing connection attempts will be rejected immediately.
func (s *Server) Close() error {
	s.Closed.Store(true)
	err := s.Listener.Close()
	if err != nil {
		return err
	}
	return nil
}

// listen is the internal loop responsible for accepting new TCP connections.
func (s *Server) listen(handler Handler) {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.Closed.Load() {
				return
			}
			log.Println(err)
			continue
		}
		go s.handle(conn, handler)
	}
}

// handle manages the lifecycle of a single connection:
// parsing the request, invoking the handler, and closing the connection.
func (s *Server) handle(conn net.Conn, handler Handler) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println(err)
		return
	}

	writer := response.NewWriter(conn)
	handler(writer, req)
}
