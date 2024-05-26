package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"log"
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

func handleConnections(c net.Conn) {
	defer c.Close()

	buffer := make([]byte, 1024)
	_, err := c.Read(buffer)
	if err != nil {
		log.Println(err)
		return
	}

	req := ParseRequest(buffer)

	fmt.Println(req)

	switch {
	case req.Target == "/":
		handleResponse(c, "HTTP/1.1 200 OK\r\n\r\n")
		return

	case strings.HasPrefix(req.Target, "/echo/"):
		v, ok := req.Headers["accept-encoding"]
		msg := strings.Split(req.Target, "/")[2]

		if ok {
			if strings.Contains(v, "gzip") {
				buffer := new(bytes.Buffer)
				compressor := gzip.NewWriter(buffer)
				compressor.Write([]byte(msg))
				compressor.Close()

				response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(buffer.String()), buffer.String())
				handleResponse(c, response)
				return
			}
		}

		//fmt.Println(msg)
		var response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(msg), msg)

		handleResponse(c, response)
		return
	case req.Target == "/user-agent":
		var response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(req.Headers["user-agent"]), req.Headers["user-agent"])

		handleResponse(c, response)
		return

	case strings.HasPrefix(req.Target, "/files/"):
		filename := strings.Split(req.Target, "/")[2]
		path := path.Join(*dir, filename)

		switch req.Method {

		case "GET":
			contents, err := os.ReadFile(path)
			if err != nil {
				fmt.Println("File NOT FOUND: \n\tfilename:", filename, "\n\tdirectory:", path)
				log.Println(err)
				c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}

			fmt.Println("file exists, reading file:", path)
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(contents), string(contents))
			handleResponse(c, response)
			return

		case "POST":
			if len(filename) == 0 {
				//TODO: Use the correct status for bad input
				c.Write([]byte("HTTP/1.1 503 Internal Server Error\r\n\r\n Filename cannot be empty"))
				return
			}

			err := os.WriteFile(path, []byte(req.Body), 0644)
			if err != nil {
				log.Println(err)
				c.Write([]byte("HTTP/1.1 503 Internal Server Error\r\n"))
				return
			}
			handleResponse(c, "HTTP/1.1 201 Created\r\n\r\n")
			return
		}

	default:
		handleResponse(c, "HTTP/1.1 404 Not Found\r\n\r\n")
		return
	}
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
		r.Headers[strings.ToLower(h[0])] = h[1]
	}

	return r
}

func handleResponse(c net.Conn, s string) {
	_, err := c.Write([]byte(s))
	if err != nil {
		log.Println(err)
	}
}

// func notFoundResponse(c net.Conn) error {
// 	return handleResponse(c, "HTTP/1.1 404 Not Found\r\n\r\n")
// }
