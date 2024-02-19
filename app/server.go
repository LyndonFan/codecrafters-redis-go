package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn) // Use a goroutine for each connection
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 0, 4096) // Initialize the buffer

	for {
		_, err := conn.Read(buffer[:cap(buffer)])
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}

		// We're hardcoding response to PONG for this stage, and ignoring the actual PING command content.
		fmt.Println("Executing: PING")
		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			return
		}
		// No need to clear the buffer since we are responding to every read regardless of its content
	}
}
