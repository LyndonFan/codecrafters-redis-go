package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

var port int = 6379

func init() {
	flag.IntVar(&port, "port", 6379, "port to listen to")
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
		os.Exit(1)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		data := make([]byte, 1024)
		dataSize, err := conn.Read(data)
		data = data[:dataSize]
		if err == io.EOF {
			fmt.Println("End of file reached")
			break
		}
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			break
		}
		fmt.Println("Received: ", data)

		// process data
		tokens, err := parseInput(string(data))
		if err != nil {
			fmt.Println("Error parsing input: ", err.Error())
			break
		}
		response, err := runTokens(tokens)
		if err != nil {
			fmt.Println("Error running tokens: ", err.Error())
			break
		}
		fmt.Println("Response: ", strings.Replace(response.Value(), TERMINATOR, "\\r\\n", -1))

		// send response
		_, err = conn.Write([]byte(response.Value()))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			break
		}
	}
}
