package main

import (
	"bufio"
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

	reader := bufio.NewReader(conn)
	var buffer strings.Builder

	for {
		data, err := reader.ReadString('\r')
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading from connection: ", err.Error())
			}
			break
		}
		buffer.WriteString(data)

		_, err = reader.Discard(1) // Discard the '\n' that follows '\r'
		if err != nil {
			fmt.Println("Error discarding from connection: ", err.Error())
			break
		}

		completeCmd := buffer.String()
		buffer.Reset()

		if completeCmd == "PING\r" {
			_, err = conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				break
			}
		}
	}
}
