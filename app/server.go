package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

var dir string

const (
	Response200 = "HTTP/1.1 200 OK\r\n\r\n"
	Response404 = "HTTP/1.1 404 Not Found\r\n\r\n"
)

type HeaderMap map[string]string

type HTTPRequest struct {
	Method      string
	Path        string
	HTTPVersion string
	Headers     map[string]string
	Body        []byte
}

func parseHeaders(content []string) map[string]string {
	headers := map[string]string{}
	for _, line := range content {
		if line == "" {
			break
		}
		parts := strings.Split(line, ": ")
		headers[strings.ToLower(parts[0])] = parts[1]
	}
	return headers
}

func parseHTTPRequest(content []byte) *HTTPRequest {
	req := &HTTPRequest{}
	byteParts := bytes.Split(content, []byte("\r\n\r\n"))
	req.Body = byteParts[1]
	parts := strings.Split(string(byteParts[0]), "\r\n")
	startLineParts := strings.Split(parts[0], " ")
	req.Method = startLineParts[0]
	req.Path = startLineParts[1]
	req.HTTPVersion = startLineParts[2]
	req.Headers = parseHeaders(parts[1:])
	return req
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	data := make([]byte, 1024)
	_, err := conn.Read(data)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		os.Exit(1)
	}
	req := parseHTTPRequest(data)
	if req.Path == "/" {
		_, err := conn.Write([]byte(Response200))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
		return
	}
	if strings.HasPrefix(req.Path, "/echo") {
		randomStr := strings.TrimPrefix(req.Path, "/echo/")
		response := "HTTP/1.1 200 OK\r\n"
		response += "Content-Type: text/plain\r\n"
		response += fmt.Sprintf("Content-Length: %d\r\n", len(randomStr))
		response += "\r\n"
		response += fmt.Sprintf("%s\r\n", randomStr)
		_, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
		}
		return
	}
	if req.Path == "/user-agent" {
		userAgent := req.Headers["user-agent"]
		response := "HTTP/1.1 200 OK\r\n"
		response += "Content-Type: text/plain\r\n"
		response += fmt.Sprintf("Content-Length: %d\r\n", len(userAgent))
		response += "\r\n"
		response += fmt.Sprintf("%s\r\n", userAgent)
		_, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
		}
		return
	}
	if strings.HasPrefix(req.Path, "/files") {
		fileName := strings.TrimPrefix(req.Path, "/files/")
		filePath := fmt.Sprintf("%s/%s", dir, fileName)
		switch req.Method {
		case "GET":
			file, err := os.Open(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					_, err = conn.Write([]byte(Response404))
					if err != nil {
						fmt.Println("Error writing to client: ", err.Error())
					}
					return
				}
				fmt.Println("Error opening file for reading: ", err.Error())
				return
			}
			fileContent, err := io.ReadAll(file)
			if err != nil {
				fmt.Println("Error reading from file: ", err.Error())
				return
			}
			response := "HTTP/1.1 200 OK\r\n"
			response += "Content-Type: application/octet-stream\r\n"
			response += fmt.Sprintf("Content-Length: %d\r\n", len(string(fileContent)))
			response += fmt.Sprintf("\r\n%s\r\n", string(fileContent))
			_, err = conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
			}
			return
		case "POST":
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Println("Error creating file: ", err.Error())
				return
			}
			defer file.Close()
			var buf bytes.Buffer
			for _, byteData := range req.Body {
				if byteData == 0 {
					break
				}
				buf.WriteByte(byteData)
			}
			_, err = buf.WriteTo(file)
			if err != nil {
				fmt.Println("Error writing content to file: ", err.Error())
				return
			}
			response := "HTTP/1.1 201 Created\r\n\r\n"
			_, err = conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
			}
			return
		}
	}
	_, err = conn.Write([]byte(Response404))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		return
	}
}

func main() {
	directory := flag.String("directory", "", "")
	flag.Parse()
	dir = *directory
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			return
		}
		go handleConnection(conn)
	}
}
