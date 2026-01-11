package server

import (
	"bytes"
	"fmt"
	"goHttp/internal/request"
	"goHttp/internal/response"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError
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

func (h *HandlerError) writeError(w io.Writer) error {
	err := response.WriteStatusLine(w, response.StatusCode(h.StatusCode))
	if err != nil {
		return err
	}
	headers := response.GetDefaultHeaders(len(h.Message))
	response.WriteHeaders(w, headers)

	_, err = w.Write([]byte(h.Message))
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

	buf := &bytes.Buffer{}
	handlerErr := handler(buf, req)
	if handlerErr != nil {
		err := handlerErr.writeError(conn)
		if err != nil {
			log.Println(err)
		}
		return
	}
	response.WriteStatusLine(conn, response.OK)
	headers := response.GetDefaultHeaders(buf.Len())
	response.WriteHeaders(conn, headers)
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Println(err)
	}
}
