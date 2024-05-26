package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

type Request struct {
	Method, Target string
	Headers        map[string]string
	Body           string
}

var dir *string

func main() {
	dir = flag.String("directory", "", "pass absolute path of the directory to get files from")
	flag.Parse()

	l, err := net.Listen("tcp", "localhost:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Println("Listening on:", l.Addr())

	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnections(conn)
	}
}

func handleConnections(c net.Conn) error {
	defer c.Close()

	buffer := make([]byte, 1024)
	_, err := c.Read(buffer)
	if err != nil {
		return err
	}

	req := ParseRequest(buffer)

	fmt.Println(req)

	switch {
	case req.Target == "/":
		return handleResponse(c, "HTTP/1.1 200 OK\r\n\r\n")

	case strings.HasPrefix(req.Target, "/echo/"):
		msg := strings.Split(req.Target, "/")[2]
		//fmt.Println(msg)
		var response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(msg), msg)

		return handleResponse(c, response)

	case req.Target == "/user-agent":
		var response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(req.Headers["User-Agent"]), req.Headers["User-Agent"])

		return handleResponse(c, response)

	case strings.HasPrefix(req.Target, "/files/"):
		filename := strings.Split(req.Target, "/")[2]
		path := path.Join(*dir, filename)

		switch req.Method {

		case "GET":
			contents, err := os.ReadFile(path)
			if err != nil {
				fmt.Println("File NOT FOUND: \n\tfilename:", filename, "\n\tdirectory:", path)
				c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return err
			}

			fmt.Println("file exists, reading file:", path)
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(contents), string(contents))
			return handleResponse(c, response)

		case "POST":
			err := os.WriteFile(path, []byte(req.Body), 0644)
			if err != nil {
				c.Write([]byte("HTTP/1.1 503 Internal Server Error\r\n"))
				return err
			}
			return handleResponse(c, "HTTP/1.1 201 Created\r\n\r\n")
		}

	default:
		return handleResponse(c, "HTTP/1.1 404 Not Found\r\n\r\n")
	}
	return nil
}

func ParseRequest(b []byte) Request {
	raw_req_string := string(b)
	r := Request{}
	r.Headers = make(map[string]string)

	req_parts := strings.Split(raw_req_string, "\r\n\r\n") // headers end with /r/n/r/n

	body_without_null := bytes.Trim([]byte(req_parts[1]), "\x00")
	r.Body = string(body_without_null)

	req_lines := strings.Split(req_parts[0], "\r\n")
	req_line_1 := strings.Split(req_lines[0], " ")

	r.Method = req_line_1[0]
	r.Target = req_line_1[1]

	raw_headers := req_lines[1:]

	for _, v := range raw_headers {
		h := strings.Split(v, ": ")
		r.Headers[h[0]] = h[1]
	}

	return r
}

func handleResponse(c net.Conn, s string) error {
	_, err := c.Write([]byte(s))
	return err
}

// func notFoundResponse(c net.Conn) error {
// 	return handleResponse(c, "HTTP/1.1 404 Not Found\r\n\r\n")
// }
