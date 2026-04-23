package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Request represents a parsed HTTP request
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

// Response represents an HTTP response to write
type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

// Listen starts a raw TCP server on the given address
func Listen(addr string, handler func(Request) Response) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer ln.Close()

	fmt.Printf("Listening on %s\n", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("accept error: %v\n", err)
			continue
		}
		// Each connection gets its own goroutine - this is what net/http does too
		go handleConn(conn, handler)
	}
}

func handleConn(conn net.Conn, handler func(Request) Response) {
	defer conn.Close()

	req, err := parseRequest(conn)
	if err != nil {
		writeResponse(conn, Response{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       []byte("bad request"),
		})
		return
	}

	resp := handler(req)
	writeResponse(conn, resp)
}

func parseRequest(conn net.Conn) (Request, error) {
	reader := bufio.NewReader(conn)

	// Parse request line: "GET /path HTTP/1.1"
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return Request{}, fmt.Errorf("read request line: %w", err)
	}
	requestLine = strings.TrimSpace(requestLine)
	parts := strings.SplitN(requestLine, " ", 3)
	if len(parts) != 3 {
		return Request{}, fmt.Errorf("malformed request line: %q", requestLine)
	}

	req := Request{
		Method:  parts[0],
		Path:    parts[1],
		Headers: make(map[string]string),
	}

	// Parse headers until blank line
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return Request{}, fmt.Errorf("read header: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // blank line = end of headers
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colonIdx])
		val := strings.TrimSpace(line[colonIdx+1:])
		req.Headers[strings.ToLower(key)] = val
	}

	// TODO: parse Content-Length and read body bytes
	// Hint: req.Headers["content-length"] gives you the length as a string

	return req, nil
}

func writeResponse(conn net.Conn, resp Response) {
	// Write status line
	fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n", resp.StatusCode, resp.StatusText)

	// Write headers
	for k, v := range resp.Headers {
		fmt.Fprintf(conn, "%s: %s\r\n", k, v)
	}
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(resp.Body))
	fmt.Fprintf(conn, "\r\n") // blank line = end of headers

	// Write body
	conn.Write(resp.Body)
}
