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
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		data := make([]byte, 2048)
		dataSize, err := conn.Read(data)
		data = data[:dataSize]
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			break
		}
		fmt.Println("Received: ", data)
		fmt.Println(strings.Contains(string(data), "\n"))
		commands := strings.Split(string(data), "\n")
		fmt.Println("Executing: ", commands)
		for _, command := range commands {
			if command == "" {
				break
			}
			fmt.Println("Executing: ", command)
			_, err = conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				break
			}
		}
		conn.Close()
	}
}
