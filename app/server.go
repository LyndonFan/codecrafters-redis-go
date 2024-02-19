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
			os.Exit(1)
		}
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 0, 4096) // Zero length but with a capacity (may change according to needs)
	tmp := make([]byte, 256) // a temporary buffer to store read chunks
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}
		buffer = append(buffer, tmp[:n]...)
		for {
			cmd, rest, found := parseCommand(buffer)
			if !found {
				break
			}
			buffer = rest // use the rest of the buffer for next commands
			if cmd == "PING" {
				_, err = conn.Write([]byte("+PONG\r\n"))
				if err != nil {
					fmt.Println("Error writing to connection:", err.Error())
					return
				}
			}
		}
	}
}

// parseCommand will parse the buffer for a Redis command terminated by \r\n
func parseCommand(buffer []byte) (cmd string, rest []byte, found bool) {
	index := -1
	for i := 0; i < len(buffer)-1; i++ {
		if buffer[i] == '\r' && buffer[i+1] == '\n' {
			index = i
			break
		}
	}
	if index == -1 { // Command not complete
		return "", buffer, false
	}
	cmd = string(buffer[:index]) // We have got a complete command, omitting the \r\n
	if cmd == "*1\r\n$4\r\nPING" {
		cmd = "PING"
	}
	rest = buffer[index+2:] // Return the rest of the buffer, skipping the \r\n of the command
	return cmd, rest, true
}
