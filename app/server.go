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
		var agent string
		for _, v := range req.Headers {
			if v.Name == "User-Agent" {
				agent = v.Value
			}
		}

		var response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(agent), agent)

		conn.Write([]byte(response))
	default:
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

	}
}

type Request struct {
	Method, Target string
	Headers        []Header
	Body           string
}

type Header struct {
	Name, Value string
}

func ParseRequest(b []byte) Request {
	raw_req_string := string(b)
	r := Request{}

	req_parts := strings.Split(raw_req_string, "\r\n\r\n")

	r.Body = req_parts[1]

	req_lines := strings.Split(req_parts[0], "\r\n")
	req_line_1 := strings.Split(req_lines[0], " ")
	r.Method = req_line_1[0]
	r.Target = req_line_1[1]

	raw_headers := req_lines[1:]

	headers := make([]Header, 0)

	for _, v := range raw_headers {
		h := strings.Split(v, ": ")
		headers = append(headers, Header{h[0], h[1]})
	}

	r.Headers = headers
	return r

	//fmt.Println(req_line[0], ",", req_line[1])
	// fmt.Println("\n1:", req_line, "\n2:", headers, "\n3:", body)
	// fmt.Println("\n1:", headers[0], "\n2:", headers[1], "\n3:", headers[2])
	// fmt.Printf("%T", headers)
}
