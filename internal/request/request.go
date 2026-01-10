package request

import (
	"bytes"
	"fmt"
	"goHttp/internal/headers"
	"io"
	"strings"
	"unicode"
)

type parserState int

const (
	StateInit parserState = iota
	requestStateParsingHeaders
	StateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

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

	return request, nil
}

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
				return 0, err
			}
			if !done && n == 0 {
				return consumed, nil
			}
			consumed += n
			if done {
				r.state = StateDone
			}
		case StateDone:
			return consumed, nil
		}
	}
}

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
