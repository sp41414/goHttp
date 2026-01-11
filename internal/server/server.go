package server

import (
	"fmt"
	"log"
	// "goHttp/internal/request"
	"net"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	Closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		Listener: listener,
	}
	go s.listen()
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

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.Closed.Load() {
				return
			}
			log.Println(err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	// req, err := request.RequestFromReader(conn)
	// if err != nil {
	// 	log.Println(err)
	// }
	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!"))
}
