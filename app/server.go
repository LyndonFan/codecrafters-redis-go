package main

import (
	"fmt"
	"net"
	"os"
	"strings"
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
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 0, 4096) // to store data received from client

	for {
		tmp := make([]byte, 256) // a temporary buffer to receive data
		n, err := conn.Read(tmp)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				return // Graceful exit on EOF
			}
			fmt.Println("Error reading from connection:", err.Error())
			return
		}
		buffer = append(buffer, tmp[:n]...)

		for {
			cmd, rest, found := parseCommand(buffer)
			if !found {
				break // Not a full command yet, go back to reading
			}
			buffer = rest // use the rest of the buffer for next commands
			if cmd == "PING" {
				_, writeErr := conn.Write([]byte("+PONG\r\n"))
				if writeErr != nil {
					fmt.Println("Error writing to connection:", writeErr.Error())
					return
				}
			}
		}
	}
}

func parseCommand(buffer []byte) (cmd string, rest []byte, found bool) {
	index := -1 // -1 indicates we haven't found the end of a command line
	// Iterate until we find \r\n in the buffer
	for i := 0; i < len(buffer)-1; i++ {
		if buffer[i] == '\r' && buffer[i+1] == '\n' {
			index = i
			break
		}
	}

	if index == -1 { // Haven't received complete command yet
		return "", buffer, false
	}

	cmd = string(buffer[:index]) // Convert the command portion of buffer to string
	rest = buffer[index+2:]      // The rest of the buffer after \r\n

	// Make sure we recognize PING according to RESP protocol
	if strings.HasPrefix(cmd, "*1\r\n$4\r\nPING") {
		return "PING", rest, true
	}

	return "", buffer, false // Default case, didn't receive a full command
}
