package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()
	conn, err := l.Accept()

	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buffer := make([]byte, 1024)
	conn.Read(buffer)

	// type Request struct {
	// 	method, target, http_version string
	// }

	//fmt.Println(string(buffer))
	req := ParseRequest(buffer)

	fmt.Println(req)

	switch {

	case req.Target == "/":
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	case strings.HasPrefix(req.Target, "/echo/"):
		msg := strings.Split(req.Target, "/")[2]
		//fmt.Println(msg)
		var response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(msg), msg)

		conn.Write([]byte(response))

	case req.Target == "/user-agent":
		var response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(req.Headers["User-Agent"]), req.Headers["User-Agent"])

		conn.Write([]byte(response))
	default:
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

	}
}

type Request struct {
	Method, Target string
	Headers        map[string]string
	Body           string
}

func ParseRequest(b []byte) Request {
	raw_req_string := string(b)
	r := Request{}
	r.Headers = make(map[string]string)

	req_parts := strings.Split(raw_req_string, "\r\n\r\n")

	r.Body = req_parts[1]

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
