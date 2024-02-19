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

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 0, 4096)

	for {
		temp := make([]byte, 2048)
		dataSize, err := conn.Read(temp)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}

		buffer = append(buffer, temp[:dataSize]...)
		for hasFullPingCommand(buffer) {
			fmt.Println("Executing: PING")
			_, err = conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				return
			}
			buffer = consumePingCommand(buffer)
		}
	}
}

func hasFullPingCommand(buffer []byte) bool {
	return len(buffer) >= 7 && string(buffer[:7]) == "ping\r\n"
}

func consumePingCommand(buffer []byte) []byte {
	return buffer[7:]
}
