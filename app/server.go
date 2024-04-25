package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	Response200 = "HTTP/1.1 200 OK\r\n\r\n"
	Response404 = "HTTP/1.1 404 Not Found\r\n\r\n"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	data := make([]byte, 1024)
	_, err = conn.Read(data)
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
	}
	_, err = conn.Write([]byte(Response404))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}
