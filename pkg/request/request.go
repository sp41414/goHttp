// Package request provides a streaming parser for HTTP/1.1 requests.
// It handles the transition from the initial request line through headers
// and into the message body.
package request

import (
	"bytes"
	"fmt"
	"github.com/sp41414/goHttp/pkg/headers"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// parserState represents the current phase of the HTTP request parsing process.
type parserState int

const (
	// StateInit is the starting state before the Request Line has been parsed.
	StateInit parserState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	// StateDone indicates the request has been fully parsed, including the body.
	StateDone
)

// Request represents a complete or partially parsed HTTP request.
type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	// state tracks the internal progress of the parser.
	state parserState
}

// RequestLine contains the metadata parsed from the first line of an HTTP request.
type RequestLine struct {
	HttpVersion   string // e.g., "1.1"
	RequestTarget string // e.g., "/index.html"
	Method        string // e.g., "GET"
}

// RequestFromReader reads from an io.Reader and returns a fully parsed Request.
// It manages the internal buffer and continues reading until the request is
// complete or an error occurs.
//
// If a Content-Length header is present, it ensures the body matches that length
// before returning successfully.
func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state: StateInit,
	}

	buf := make([]byte, 8)
	readToIndex := 0
	for {
		if request.state == StateDone {
			break
		}
		if readToIndex >= len(buf) {
			dt := make([]byte, len(buf)*2)
			copy(dt, buf)
			buf = dt
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("Error: could not read request (%v)", err)
		}

		readToIndex += n

		read, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("Error: could not parse request (%v)", err)
		}

		copy(buf, buf[read:readToIndex])
		readToIndex -= read
	}

	if request.state == StateInit {
		return nil, fmt.Errorf("Error: found EOF before end of request line")
	}

	contentLengthHeaders := request.Headers.Get("Content-Length")
	if contentLengthHeaders != "" {
		clh, err := strconv.Atoi(contentLengthHeaders)
		if err != nil {
			return nil, fmt.Errorf("Error: could not parse Content-Length")
		}
		cl := len(request.Body)
		if clh != cl {
			return nil, fmt.Errorf("Error: found EOF before length body %d is the same as Content-Length %d\n", cl, clh)
		}
	}

	return request, nil
}

// parse processes a slice of bytes and updates the request state.
// It returns the number of bytes consumed. This is useful for incremental
// parsing where data might arrive in chunks.
func (r *Request) parse(data []byte) (int, error) {
	consumed := 0
	for {
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(data[consumed:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return 0, nil
			}
			r.RequestLine = *rl
			consumed += n
			r.state = requestStateParsingHeaders
		case requestStateParsingHeaders:
			if r.Headers == nil {
				h := headers.NewHeaders()
				r.Headers = h
			}
			n, done, err := r.Headers.Parse(data[consumed:])
			if err != nil {
				return consumed, err
			}
			if !done && n == 0 {
				return consumed, nil
			}
			consumed += n
			if done {
				r.state = requestStateParsingBody
			}
		case requestStateParsingBody:
			clHeader := r.Headers.Get("Content-Length")
			if clHeader == "" {
				r.state = StateDone
				return consumed, nil
			}

			contentLength, err := strconv.Atoi(clHeader)
			if err != nil {
				return consumed, fmt.Errorf("invalid body: content-length is not a number (%v)", err)
			}

			n := len(data[consumed:])
			r.Body = append(r.Body, data[consumed:consumed+n]...)
			consumed += n
			if contentLength == len(r.Body) {
				r.state = StateDone
			}
			return consumed, nil
		case StateDone:
			return consumed, nil
		}
	}
}

// parseRequestLine extracts the Method, RequestTarget, and HttpVersion from the
// first line of a request. It expects the line to end with \r\n.
func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}
	read := idx + len("\r\n")

	splitRequestLine := bytes.Split(data[:idx], []byte(" "))

	if len(splitRequestLine) != 3 {
		return nil, 0, fmt.Errorf("invalid request line, request line must be in Method RequestTarget HttpVersion format")
	}

	method, requestTarget, httpVersion := string(splitRequestLine[0]), string(splitRequestLine[1]), string(splitRequestLine[2])

	if strings.ToUpper(method) != method {
		return nil, 0, fmt.Errorf("invalid request line, method name must be in full capital letters")
	}
	for _, l := range method {
		if !unicode.IsLetter(l) {
			return nil, 0, fmt.Errorf("invalid request line, method name must be alphabetical")
		}
	}

	versionParts := strings.Split(httpVersion, "/")
	if len(versionParts) != 2 || versionParts[0] != "HTTP" {
		return nil, 0, fmt.Errorf("invalid request line, please ensure http version is HTTP/1.1")
	}

	version := versionParts[1]
	if version != "1.1" {
		return nil, 0, fmt.Errorf("invalid request line, please ensure http version is HTTP/1.1")
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, read, nil
}
