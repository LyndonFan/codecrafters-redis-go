package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

var port int

type ReplicationInfo struct {
	Host string
	Port int
}

var replicationInfo ReplicationInfo

func (ri ReplicationInfo) String() string {
	if ri.IsMaster() {
		return "{master}"
	}
	return fmt.Sprintf("{%s:%d}", ri.Host, ri.Port)
}

func (ri ReplicationInfo) IsMaster() bool {
	return ri.Host == "" && ri.Port == 0
}

func init() {
	flag.IntVar(&port, "port", 6379, "port to listen to")
	flag.StringVar(&(replicationInfo.Host), "replicaof", "", "if specified, the host and port of its master")
	flag.Parse()
	if replicationInfo.Host != "" {
		if len(flag.Args()) == 0 {
			fmt.Printf("Missing port\n")
			os.Exit(1)
		}
		replicationPort, err := strconv.Atoi(flag.Args()[0])
		if err != nil {
			fmt.Printf("Invalid port: %s\n", flag.Args()[0])
			os.Exit(1)
		}
		replicationInfo.Port = replicationPort
	}
}

func main() {
	fmt.Printf("Replication info: %v\n", replicationInfo)
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
		tokens, err := token.ParseInput(string(data))
		if err != nil {
			fmt.Println("Error parsing input: ", err.Error())
			break
		}
		response, err := runTokens(tokens)
		if err != nil {
			response = &token.Token{Type: token.ErrorType, SimpleValue: fmt.Sprintf("error: %s", err.Error())}
		}
		fmt.Println("Response: ", strings.Replace(response.EncodedString(), token.TERMINATOR, "\\r\\n", -1))

		// send response
		_, err = conn.Write([]byte(response.EncodedString()))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			break
		}
	}
}
