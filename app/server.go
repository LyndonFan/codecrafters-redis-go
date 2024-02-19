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
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			break
		}

		for {
			buffer := make([]byte, 1024)
			_, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Error reading from connection or connection closed by client: ", err.Error())
				break
			}
			_, err = conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				break
			}
		}

		conn.Close()
	}
}
