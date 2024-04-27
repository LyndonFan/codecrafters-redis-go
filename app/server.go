package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var port int

type ReplicationRole struct {
	Host string
	Port int
}

func (m *ReplicationRole) IsMaster() bool {
	return m.Host == "" && m.Port == 0
}

func (m *ReplicationRole) String() string {
	if m.IsMaster() {
		return ""
	}
	return fmt.Sprintf("%s:%d", m.Host, m.Port)
}

func (m *ReplicationRole) Set(value string) error {
	parts := strings.Split(value, " ")
	if len(parts) != 2 {
		return fmt.Errorf("invalid replication master: %s", value)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	m.Host = parts[0]
	m.Port = port
	return nil
}

var role ReplicationRole

func init() {
	flag.IntVar(&port, "port", 6379, "port to listen to")
	flag.Var(&role, "replicaof", "if specified, the host and port of its master")
	flag.Parse()
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
		os.Exit(1)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		data := make([]byte, 1024)
		dataSize, err := conn.Read(data)
		data = data[:dataSize]
		if err == io.EOF {
			fmt.Println("End of file reached")
			break
		}
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			break
		}
		fmt.Println("Received: ", data)

		// process data
		tokens, err := parseInput(string(data))
		if err != nil {
			fmt.Println("Error parsing input: ", err.Error())
			break
		}
		response, err := runTokens(tokens)
		if err != nil {
			response = &Token{Type: errorType, SimpleValue: fmt.Sprintf("error: %s", err.Error())}
		}
		fmt.Println("Response: ", strings.Replace(response.Value(), TERMINATOR, "\\r\\n", -1))

		// send response
		_, err = conn.Write([]byte(response.Value()))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			break
		}
	}
}
