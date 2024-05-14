package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
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

	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		os.Exit(1)
	}

	data := string(buff[:n])

	lines := strings.Split(data, "\r\n")

	words := strings.Split(lines[0], " ")
	target := words[1]
	parts := strings.Split(target, "/")

	response := []byte{}
	if target == "/" {
		response = ([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if len(parts) > 2 && parts[1] == "echo" {
		response = ([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(len(parts[2])) + "\r\n\r\n" + parts[2]))
	} else {
		response = ([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}

	_, err = conn.Write(response)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		os.Exit(1)
	}

	conn.Close()

}
