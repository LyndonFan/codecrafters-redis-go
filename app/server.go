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
            continue
        }

        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    buffer := make([]byte, 0, 4096)
    temp := make([]byte, 2048)

    for {
        dataSize, err := conn.Read(temp)
        if err != nil {
            fmt.Println("Error reading from connection: ", err.Error())
            return
        }

        buffer = append(buffer, temp[:dataSize]...)
        commands := string(buffer)

        // Check if we received at least one full command
        for {
            if len(commands) == 0 {
                break
            }

            endOfCommand := -1
            for i, ch := range commands {
                if ch == '\n' {
                    if i > 0 && commands[i-1] == '\r' {
                        endOfCommand = i + 1
                        break
                    }
                }
            }

            if endOfCommand == -1 {
                // We don't have a full command yet, so wait for more data
                break
            }

            command := commands[:endOfCommand]
            commands = commands[endOfCommand:]
            buffer = buffer[endOfCommand:]

            fmt.Println("Executing: ", command)
            _, err = conn.Write([]byte("+PONG\r\n"))
            if err != nil {
                fmt.Println("Error writing to connection: ", err.Error())
                return
            }
        }
    }
}
