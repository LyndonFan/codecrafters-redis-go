package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/token"
)

var port int

var repl *replication.Replicator

func init() {
	flag.IntVar(&port, "port", 6379, "port to listen to")
	var replHost string
	flag.StringVar(&replHost, "replicaof", "", "if specified, the host and port of its master")
	flag.Parse()
	remainingArgs := flag.Args()
	var err error
	if len(remainingArgs) == 0 {
		repl, err = replication.GetReplicator(port, "", "")
	} else if replHost != "" {
		repl, err = replication.GetReplicator(port, replHost, remainingArgs[0])
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	fmt.Printf("Replication info: %v\n", repl)
	fmt.Println("Logs from your program will appear here!")

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
		os.Exit(1)
	}
	defer listener.Close()
	masterConn, err := repl.HandshakeWithMaster()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if masterConn != nil {
		go handleConnection(masterConn, true)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error: ", err.Error())
			continue
		}
		go handleConnection(conn, false)
	}
}

func handleConnection(conn net.Conn, fromMaster bool) {
	fmt.Println("Received connection from", conn.RemoteAddr().String())
	defer conn.Close()
	var data []byte
	if fromMaster {
		// handle RDB file
		data = make([]byte, 1024)
		dataSize, err := conn.Read(data)
		if err != nil {
			if err == io.EOF {
				fmt.Println("End of file reached")
			} else {
				fmt.Println("Error reading from connection: ", err.Error())
			}
		} else {
			data = data[:dataSize]
			fmt.Println("Received RDB file: ", strings.Replace(string(data), token.TERMINATOR, "\\r\\n", -1))
		}
	}
	for {
		data = make([]byte, 1024)
		dataSize, err := conn.Read(data)
		data = data[:dataSize]
		if err != nil {
			if err == io.EOF {
				fmt.Println("End of file reached")
			} else {
				fmt.Println("Error reading from connection: ", err.Error())
			}
			break
		}
		fmt.Println("Received: ", strings.Replace(string(data), token.TERMINATOR, "\\r\\n", -1))

		// process data
		err = repl.HandshakeWithFollower(conn, data)
		if err == nil {
			continue
		}

		tokens, err := token.ParseInput(string(data))
		var responses []*token.Token
		if err != nil {
			err = fmt.Errorf("error parsing input: %v", err)
			fmt.Println(err)
			responses = []*token.Token{token.TokeniseError(err)}
		} else {
			fmt.Println("Tokens:")
			for _, t := range tokens {
				fmt.Println(*t)
			}
			responses, err = runTokens(tokens)
			if err != nil {
				fmt.Println(err)
				responses = []*token.Token{token.TokeniseError(err)}
			}
		}
		for _, response := range responses {
			fmt.Println("Response: ", strings.Replace(response.EncodedString(), token.TERMINATOR, "\\r\\n", -1))
			if fromMaster {
				fmt.Println("Not sending response to master")
			} else {
				_, err = conn.Write([]byte(response.EncodedString()))
				if err != nil {
					fmt.Println("Error writing to connection: ", err.Error())
					break
				}
			}
		}
	}
}
