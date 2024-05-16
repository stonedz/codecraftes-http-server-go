package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Parse command line arguments
	directory := flag.String("directory", ".", "Directory to serve files from")
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

		go handleConnection(conn, directory)
	}

}

func handleConnection(conn net.Conn, directory *string) {
	defer conn.Close()
	fmt.Println("Handling connection...")

	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		os.Exit(1)
	}

	data := string(buff[:n])

	lines := strings.Split(data, "\r\n")

	words := strings.Split(lines[0], " ")
	verb := words[0]
	target := words[1]
	req_parts := strings.Split(target, "/")
	user_agent := getHeader(lines, "User-Agent")
	body := lines[len(lines)-1]

	response := []byte{}
	switch verb {
	case "GET":
		response = handleGetRequest(req_parts, user_agent, directory, target)
	case "POST":
		response = handlePostRequest(req_parts, directory, body)
	}

	_, err = conn.Write(response)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		os.Exit(1)
	}

}

func handleGetRequest(req_parts []string, user_agent string, directory *string, target string) []byte {
	var response []byte
	if target == "/" {
		response = ([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if req_parts[1] == "user-agent" {
		response = ([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(len(user_agent)) + "\r\n\r\n" + user_agent))
	} else if len(req_parts) > 2 && req_parts[1] == "echo" {
		response = ([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(len(req_parts[2])) + "\r\n\r\n" + req_parts[2]))
	} else if len(req_parts) > 2 && req_parts[1] == "files" && checkFileExists(*directory+req_parts[2]) {
		file_contents := getFileContents(*directory + req_parts[2])
		response = ([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + fmt.Sprint(len(file_contents)) + "\r\n\r\n" + file_contents))
	} else {
		response = ([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	return response

}

func handlePostRequest(req_parts []string, directory *string, body string) []byte {
	var response []byte
	if len(req_parts) > 2 && req_parts[1] == "files" {
		file_name := req_parts[2]
		err := os.WriteFile(*directory+file_name, []byte(body), 0644)
		if err != nil {
			fmt.Println("Error writing file: ", err.Error())
			response = ([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		} else {
			response = ([]byte("HTTP/1.1 201 Created\r\nLocation: /files/" + file_name + "\r\n\r\n"))
		}
	}
	return response
}

func getHeader(lines []string, name string) string {
	for _, line := range lines {
		parts := strings.Split(line, ": ")
		if parts[0] == name {
			return parts[1]
		}
	}
	return ""
}

func checkFileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func getFileContents(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file: ", err.Error())
		os.Exit(1)
	}
	return string(data)
}
