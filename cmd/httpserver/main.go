package main

import (
	"fmt"
	"goHttp/internal/headers"
	"goHttp/internal/request"
	"goHttp/internal/response"
	"goHttp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func handler(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		trimmed := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
		res, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", trimmed))
		if err != nil {
			log.Println(err)
			return
		}
		w.WriteStatusLine(response.StatusCode(res.StatusCode))

		headers := headers.NewHeaders()
		for k, v := range res.Header {
			loweredK := strings.ToLower(k)
			if loweredK == "content-length" || loweredK == "transfer-encoding" {
				continue
			}

			_, err := headers.Add(loweredK, strings.Join(v, ", "))
			if err != nil {
				log.Println(err)
				return
			}
		}
		_, err = headers.Add("Transfer-Encoding", "chunked")
		if err != nil {
			log.Println(err)
			return
		}
		w.WriteHeaders(headers)

		chunk := make([]byte, 1024)
		for {
			n, err := res.Body.Read(chunk)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println(err)
				return
			}

			n, err = w.WriteChunkedBody(chunk[:n])
			if err != nil {
				log.Println(err)
				return
			}
		}

		w.WriteChunkedBodyDone()
		return
	}

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		w.WriteStatusLine(response.BAD_REQUEST)
		body := []byte(`
			<html>
			  <head>
				<title>400 Bad Request</title>
			  </head>
			  <body>
				<h1>Bad Request</h1>
				<p>Your request honestly kinda sucked.</p>
			  </body>
			</html>
		`)
		h := response.GetDefaultHeaders(len(body))
		h.OverrideValue("Content-Type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody(body)
	case "/myproblem":
		w.WriteStatusLine(response.INTERNAL_SERVER_ERROR)
		body := []byte(`
			<html>
			  <head>
				<title>500 Internal Server Error</title>
			  </head>
			  <body>
				<h1>Internal Server Error</h1>
				<p>Okay, you know what? This one is on me.</p>
			  </body>
		    </html>
		`)
		h := response.GetDefaultHeaders(len(body))
		h.OverrideValue("Content-Type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody(body)
	default:
		w.WriteStatusLine(response.OK)
		body := []byte(`
			<html>
			  <head>
				<title>200 OK</title>
			  </head>
			  <body>
				<h1>Success!</h1>
				<p>Your request was an absolute banger.</p>
			  </body>
			</html>
		`)
		h := response.GetDefaultHeaders(len(body))
		h.OverrideValue("Content-Type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody(body)
	}
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
