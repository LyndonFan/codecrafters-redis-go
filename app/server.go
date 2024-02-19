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

		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 0, 4096)

	for {
		dataSize, err := conn.Read(buffer[:cap(buffer)])
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}

		buffer = buffer[:dataSize]
		if hasFullPingCommand(buffer) {
			fmt.Println("Executing: PING")
			_, err = conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				return
			}
			// Clear the buffer for the next command
			buffer = buffer[:0]
		}
	}
}

func hasFullPingCommand(buffer []byte) bool {
	// We're considering the Redis protocol encoded PING command: *1\r\n$4\r\nPING\r\n
	return len(buffer) >= 11 && string(buffer[:11]) == "*1\r\n$4\r\nPING\r\n"
}

func consumePingCommand(buffer []byte) []byte {
	if len(buffer) > 11 {
		return buffer[11:]
	}
	return buffer[:0]
}
