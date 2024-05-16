package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

// Response struct
type Response struct {
	statusCode int
	headers    map[string]string
	body       string
}

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

func buildResponse(response *Response) string {
	statusMessage := getStatusMessage(response.statusCode)
	headers := ""
	for key, value := range response.headers {
		headers += key + ": " + value + "\r\n"
	}
	return "HTTP/1.1 " + fmt.Sprint(response.statusCode) + " " + statusMessage + "\r\n" + headers + "\r\n" + response.body
}

func getStatusMessage(statusCode int) string {
	switch statusCode {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	default:
		return "Unknown"
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

	encoding_requests := strings.Split(getHeader(lines, "Accept-Encoding"), ",")
	for i, encoding := range encoding_requests {
		encoding_requests[i] = strings.TrimSpace(encoding)
	}

	body := lines[len(lines)-1]

	var response *Response
	switch verb {
	case "GET":
		response = handleGetRequest(req_parts, user_agent, directory, target)
	case "POST":
		response = handlePostRequest(req_parts, directory, body)
	}

	response = handleCompressionEncoding(encoding_requests, response)

	_, err = conn.Write([]byte(buildResponse(response)))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		os.Exit(1)
	}

}

func handleCompressionEncoding(encoding_request []string, response *Response) *Response {
	for _, encoding := range encoding_request {
		if encoding == "gzip" {
			response.headers["Content-Encoding"] = "gzip"
			response.body = compress(response.body)
			response.headers["Content-Length"] = fmt.Sprint(len(response.body))
		}
	}

	return response
}

func compress(data string) string {
	var b strings.Builder
	gz := gzip.NewWriter(&b)
	_, err := gz.Write([]byte(data))
	if err != nil {
		fmt.Println("Error compressing data: ", err.Error())
		os.Exit(1)
	}
	err = gz.Close()
	if err != nil {
		fmt.Println("Error closing gzip writer: ", err.Error())
		os.Exit(1)
	}
	return b.String()
}

func handleGetRequest(req_parts []string, user_agent string, directory *string, target string) *Response {
	var response Response
	if target == "/" {
		response.statusCode = 200
		response.body = ""
	} else if req_parts[1] == "user-agent" {
		response.statusCode = 200
		response.headers = map[string]string{"Content-Type": "text/plain", "Content-Length": fmt.Sprint(len(user_agent))}
		response.body = user_agent
	} else if len(req_parts) > 2 && req_parts[1] == "echo" {
		response.statusCode = 200
		response.headers = map[string]string{"Content-Type": "text/plain", "Content-Length": fmt.Sprint(len(req_parts[2]))}
		response.body = req_parts[2]
	} else if len(req_parts) > 2 && req_parts[1] == "files" && checkFileExists(*directory+req_parts[2]) {
		file_contents := getFileContents(*directory + req_parts[2])
		response.statusCode = 200
		response.headers = map[string]string{"Content-Type": "application/octet-stream", "Content-Length": fmt.Sprint(len(file_contents))}
		response.body = file_contents
	} else {
		response.statusCode = 404
	}
	return &response

}

func handlePostRequest(req_parts []string, directory *string, body string) *Response {
	var response Response
	if len(req_parts) > 2 && req_parts[1] == "files" {
		file_name := req_parts[2]
		err := os.WriteFile(*directory+file_name, []byte(body), 0644)
		if err != nil {
			fmt.Println("Error writing file: ", err.Error())
			response.statusCode = 500
		} else {
			response.statusCode = 201
			response.headers = map[string]string{"Location": "/files/" + file_name}
		}
	}
	return &response
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
