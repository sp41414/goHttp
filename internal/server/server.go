package server

import (
	"fmt"
	"github.com/sp41414/goHttp/internal/request"
	"github.com/sp41414/goHttp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Handler func(w *response.Writer, req *request.Request)
type HandlerError struct {
	StatusCode int
	Message    string
}

type Server struct {
	Listener net.Listener
	Closed   atomic.Bool
}

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

func (s *Server) Close() error {
	s.Closed.Store(true)
	err := s.Listener.Close()
	if err != nil {
		return err
	}
	return nil
}

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
