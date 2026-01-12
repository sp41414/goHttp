package response

import (
	"fmt"
	"goHttp/internal/headers"
	"io"
	"strconv"
	"strings"
)

type StatusCode int

type writerState int
type Writer struct {
	inner io.Writer
	State writerState
}

const (
	StatusLine writerState = iota
	Header
	Body
	Trailers
	Done
)

const (
	OK                    StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	INTERNAL_SERVER_ERROR StatusCode = 500
)

func NewWriter(inner io.Writer) *Writer {
	return &Writer{
		State: StatusLine,
		inner: inner,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != StatusLine {
		return fmt.Errorf("Error: unexpected state, expected state to be StatusLine")
	}

	switch statusCode {
	case OK:
		_, err := w.inner.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
	case BAD_REQUEST:
		_, err := w.inner.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
	case INTERNAL_SERVER_ERROR:
		_, err := w.inner.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
	default:
		_, err := w.inner.Write([]byte(fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)))
		if err != nil {
			return err
		}
	}

	w.State = Header
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"content-length": strconv.Itoa(contentLen),
		"connection":     "close",
		"content-type":   "text/plain",
	}
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != Header {
		return fmt.Errorf("Error: unexpected state, expected state to be Header")
	}

	for k, v := range headers {
		_, err := w.inner.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}

	// blank line before the body
	_, err := w.inner.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.State = Body
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State != Body {
		return 0, fmt.Errorf("Error: unexpected state, expected state to be Body")
	}

	n, err := w.inner.Write(p)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.State != Body {
		return 0, fmt.Errorf("Error: unexpected state, expected state to be Body")
	}

	n := len(p)
	hl := fmt.Sprintf("%x", n)

	_, err := w.inner.Write([]byte(hl + "\r\n"))
	if err != nil {
		return 0, err
	}

	_, err = w.inner.Write(p)
	if err != nil {
		return 0, err
	}

	_, err = w.inner.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.State != Body {
		return 0, fmt.Errorf("Error: unexpected state, expected state to be Body")
	}

	n, err := w.inner.Write([]byte("0\r\n"))
	if err != nil {
		return 0, err
	}

	w.State = Trailers
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.State != Trailers {
		return fmt.Errorf("Error: unexpected state, expected state to be Trailers")
	}

	for k, v := range h {
		loweredK := strings.ToLower(strings.TrimSpace(k))
		_, err := w.inner.Write([]byte(fmt.Sprintf("%s: %s\r\n", loweredK, strings.TrimSpace(v))))
		if err != nil {
			return err
		}
	}

	// blank line before the end
	_, err := w.inner.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.State = Done
	return nil
}
