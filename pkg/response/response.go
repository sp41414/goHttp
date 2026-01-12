// Package response provides a stateful HTTP/1.1 response writer.
// It ensures that response components (Status Line, Headers, Body, and Trailers)
// are written in the correct order as defined by the HTTP specification.
package response

import (
	"fmt"
	"github.com/sp41414/goHttp/pkg/headers"
	"io"
	"strconv"
	"strings"
)

// StatusCode represents an HTTP response status code e.g.(200, 400, 500).
type StatusCode int

// Writer wraps an io.Writer to manage the lifecycle of an HTTP response.
// It maintains an internal state to prevent out-of-order writes.
type Writer struct {
	inner io.Writer
	State writerState
}

// writerState defines the valid stages of a response lifecycle.
type writerState int

const (
	StatusLine writerState = iota // Initial state: expectation of WriteStatusLine
	Header                        // Expecting WriteHeaders
	Body                          // Expecting WriteBody or WriteChunkedBody
	Trailers                      // Expecting WriteTrailers after chunked transfer
	Done                          // Final state: no more writes allowed
)

const (
	OK                    StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	INTERNAL_SERVER_ERROR StatusCode = 500
)

// NewWriter initializes a Writer in the StatusLine state.
func NewWriter(inner io.Writer) *Writer {
	return &Writer{
		State: StatusLine,
		inner: inner,
	}
}

// WriteStatusLine writes the HTTP/1.1 status line.
// It transitions the writer from StatusLine to Header state.
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

// GetDefaultHeaders returns a Headers map pre-populated with standard
// fields like content-length, connection: close, and text/plain content-type.
func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"content-length": strconv.Itoa(contentLen),
		"connection":     "close",
		"content-type":   "text/plain",
	}
}

// WriteHeaders writes the provided headers followed by the required
// empty line (\r\n). It transitions the writer to the Body state.
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

// WriteBody writes raw data to the inner writer.
// It should only be called after WriteHeaders.
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

// WriteChunkedBody writes a single data chunk using HTTP Chunked Transfer Encoding.
// It automatically handles the hex-length prefix and CRLF suffixes.
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

// WriteChunkedBodyDone writes the final "0" chunk to signal the end of a
// chunked response and transitions the state to allow for Trailers.
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

// WriteTrailers writes trailing headers and the final terminating empty line.
// It transitions the writer to the Done state.
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
