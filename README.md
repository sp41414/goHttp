# goHttp

This project is to understand HTTP protocol by writing it in Go.
The HTTP protocol version that was built is HTTP/1.1.

## Getting Started
1. Start a TCP Listener by running or building `./cmd/tcplistener/`
2. Start a UDP Sender by running or building `./cmd/udpsender`
3. Start an HTTP Server by running or building `./cmd/httpserver`

## Features
### HTTP Server
1. The HTTP Server allows you to write a custom handler which you can pass in the response writer, and the user's request.
2. In the custom handler you can write a request line with status codes: `200`, `400`, and `500`. these are mapped to go ENUMs `OK`, `BAD_REQUEST`, `INTERNAL_SERVER_ERROR` and you can pass in any integers too.
3. Write a custom header by first creating one using `NewHeaders()`, and then you can add values with `Add()`, Override keys with `Override()`, Override values with `OverrideValue()`
4. Write a custom body depending on the header's Content-Length and Content-Type (not validated by the internals for more user control)
5. Send the body with chunked encoding
6. Write trailers by first writing a Trailer: key1, key2, to your headers, then creating new headers and adding the same keys but with values then using the `WriteTrailers()` function.
### TCP Listener
1. The TCP Listener starts a server on port :2000 (customizable within `./cmd/tcplistener/main.go`)
2. Listens to requests and parses the request line, headers, and body. For now, it just prints it. You can do whatever you want with it.
### UDP Sender
1. The UDP Sender sends to localhost:2000 (customizable within `./cmd/udpsender/main.go`)
2. Once the UDP Sender starts, hit `CTRL + C` to stop it. it will infinitely listen to your prompt and write it to the specified listener
