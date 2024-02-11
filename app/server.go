package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var dir *string

func main() {
	fmt.Println("Logs from your program will appear here!")
	dir = flag.String("directory", "", "absolute file path")
	flag.Parse()

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
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	reader := bufio.NewReader(conn)
	defer conn.Close()

	requestLine, _ := reader.ReadString('\n')

	reader.ReadString('\n')

	userAgent, _ := reader.ReadString('\n')
	lines := strings.Fields(requestLine)
	method := lines[0]
	path := lines[1]

	var res string
	if path == "/" {
		res = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(path, "/echo") {
		dynamicString := strings.TrimPrefix(path, "/echo/")
		res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(dynamicString), dynamicString)
	} else if strings.HasPrefix(path, "/user-agent") {
		userAgent := strings.TrimRight(strings.Split(userAgent, " ")[1], "\r\n")
		res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s\r\n", len(userAgent), userAgent)
	} else if strings.HasPrefix(path, "/files") && strings.ToLower(method) == "get" {
		filePath := filepath.Join(*dir, strings.TrimPrefix(path, "/files"))
		body, err := os.ReadFile(filePath)
		if err != nil {
			res = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s\r\n", len(body), body)
		}
	} else if strings.HasPrefix(path, "/files") && strings.ToLower(method) == "post" {
		header := make(map[string]string)
		for line, err := reader.ReadString('\n'); line != "\r\n"; line, err = reader.ReadString('\n') {
			if err != nil {
				fmt.Println("Error", err.Error())
				return
			}
			lineSplit := strings.Split(line, ": ")
			header[lineSplit[0]] = strings.TrimRight(lineSplit[1], "\r\n")
		}
		filePath := filepath.Join(*dir, strings.TrimPrefix(path, "/files"))
		contentLength, err := strconv.Atoi(header["Content-Length"])
		if err != nil {
			fmt.Println("Error  ", err.Error())
			return
		}
		body := make([]byte, contentLength)
		_, err = reader.Read(body)
		if err != nil {
			fmt.Println("Error  ", err.Error())
			return
		}
		err = os.WriteFile(filePath, body, 0644)
		if err != nil {
			fmt.Println("Error writing file: ", err.Error())
			return
		}
		_, err = conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		if err != nil {
			fmt.Println("Error writing file: ", err.Error())
			return
		}
	} else {
		res = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	conn.Write([]byte(res))
}
