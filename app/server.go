package main

import (
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

func parseHeaders(content []string) HeaderMap {
	headers := HeaderMap{}
	for _, line := range content {
		if line == "" {
			break
		}
		parts := strings.Split(line, ": ")
		headers[strings.ToLower(parts[0])] = parts[1]
	}
	return headers
}

func handleConnection(conn net.Conn) {
	data := make([]byte, 1024)
	_, err := conn.Read(data)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		os.Exit(1)
	}
	parts := strings.Split(string(data), "\r\n")
	startLine := parts[0]
	path := strings.Split(startLine, " ")[1]
	if path == "/" {
		_, err := conn.Write([]byte(Response200))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
	} else if strings.HasPrefix(path, "/echo") {
		randomStr := strings.TrimPrefix(path, "/echo/")
		response := "HTTP/1.1 200 OK\r\n"
		response += "Content-Type: text/plain\r\n"
		response += fmt.Sprintf("Content-Length: %d\r\n", len(randomStr))
		response += "\r\n"
		response += fmt.Sprintf("%s\r\n", randomStr)
		_, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
	} else if path == "/user-agent" {
		headers := parseHeaders(parts[1:])
		userAgent := headers["user-agent"]
		response := "HTTP/1.1 200 OK\r\n"
		response += "Content-Type: text/plain\r\n"
		response += fmt.Sprintf("Content-Length: %d\r\n", len(userAgent))
		response += "\r\n"
		response += fmt.Sprintf("%s\r\n", userAgent)
		_, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
	} else if strings.HasPrefix(path, "/files") {
		fileName := strings.TrimPrefix(path, "/files/")
		filePath := fmt.Sprintf("%s/%s", dir, fileName)
		file, err := os.Open(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				_, err = conn.Write([]byte(Response404))
				if err != nil {
					fmt.Println("Error writing to client: ", err.Error())
					os.Exit(1)
				}
				os.Exit(0)
			}
			fmt.Println("Error opening file for reading: ", err.Error())
			os.Exit(1)
		}
		fileContent, err := io.ReadAll(file)
		if err != nil {
			fmt.Println("Error reading from file: ", err.Error())
			os.Exit(1)
		}
		response := "HTTP/1.1 200 OK\r\n"
		response += "Content-Type: application/octet-stream\r\n"
		response += fmt.Sprintf("Content-Length: %d\r\n", len(string(fileContent)))
		response += fmt.Sprintf("\r\n%s\r\n", string(fileContent))
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
	} else {
		_, err = conn.Write([]byte(Response404))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
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
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
