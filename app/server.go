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

	buffer := make([]byte, 0, 4096) // store data received from client

	for {
		tmp := make([]byte, 256) // temporary buffer to receive data
		n, err := conn.Read(tmp)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				// Don't return on EOF, as the client might send multiple commands in one connection
				break
			}
			fmt.Println("Error reading from connection:", err.Error())
			return
		}

		if n == 0 { // no new data read, go back to reading
			continue
		}

		buffer = append(buffer, tmp[:n]...)

		for {
			cmd, rest, found := parseCommand(buffer)
			if !found {
				break // not a complete command, wait for more data
			}
			buffer = rest // residual data could be part of the next command
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
	index := strings.Index(string(buffer), "\r\n")
	if index == -1 {
		return "", buffer, false // complete command not yet received
	}
	commandLine := string(buffer[:index])
	rest = buffer[index+2:]
	if strings.HasPrefix(commandLine, "*1\r\n$4\r\nPING") {
		return "PING", rest, true
	}
	return "", buffer, false // unrecognized command
}
