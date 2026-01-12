# goHttp

This project is to understand HTTP protocol by writing it in Go.
The HTTP protocol version that was built is HTTP/1.1.

## Prerequisites
- Go 1.20 or higher
- Optionally a terminal for testing e.g.(`curl` or `netcat`)

## Getting Started
1. Start a TCP Listener by running or building `./cmd/tcplistener/`
```bash
go run ./cmd/tcplistener/main.go
```
2. Start a UDP Sender by running or building `./cmd/udpsender`
```bash
go run ./cmd/udpsender/main.go
```
3. Start an HTTP Server by running or building `./cmd/httpserver`
```bash
go run ./cmd/httpserver/main.go
```

## Features
### Design
This project implements HTTP/1.1 spec through a custom State Machine. This ensures that messages are parsed and written in the correct order (Status Line -> Headers -> Body -> Trailers).

1. HTTP Server
The server package manages concurrent TCP connections, each handled in its own goroutine.
- Custom Handlers: Define logic using `func(w *response.Writer, req *request.Request)`.
- Stateful Writing: The `Writer` prevents malformed responses by enforcing the protocol order. 
- Chunked Encoding: Support for `Transfer-Encoding: chunked` with a dedicated `WriteChunkedBody` method.
- Trailers: Ability to send metadata after the body has been streamed.

2. Header Management
The `headers` package provides a case-insensitive map for managing HTTP tokens.
- `Add(key, value)`: Validates keys against RFC 9110 tokens.
- `Override(prev, new, val)`: Renames and updates existing keys.
- `Get(key)`: Case-insensitive retrieval.
3. UDP & TCP Utils
- TCP Listener: Demonstrates the `request` package's ability to parse streaming data from a raw `net.Conn`.
- UDP Sender: A CLI tool to send manual payloads to local ports for testing.

## Example Usage
Here is how you can build a simple server using the packages:
```go
package main

import (
    "github.com/sp41414/goHttp/pkg/server"
    "github.com/sp41414/goHttp/pkg/response"
    "github.com/sp41414/goHttp/pkg/request"
)

func main() {
    handler := func(w *response.Writer, req *request.Request) {
        // 1. Write Status Line
        w.WriteStatusLine(response.OK)
        
        // 2. Set Headers
        h := response.GetDefaultHeaders(13) // Content Length 13 is passed in
        w.WriteHeaders(h)
        
        // 3. Write Body
        w.WriteBody([]byte("Hello, World!"))
    }

    s, _ := server.Serve(8080, handler)
    defer s.Close()
    
    // Server runs in a background goroutine
    select {} 
}
```
You can see how I implemented it in `./cmd/httpserver/main.go` for more examples.

## Documentation
- **Online**: [pkg.go.dev/github.com/sp41414/goHttp](https://pkg.go.dev/github.com/sp41414/goHttp).
- **Local**: Run `pkgsite` on the root of the project directory to view the docs offline at http://localhost:8080.
