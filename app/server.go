package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/token"
)

var port int

var repl *replication.Replicator

func init() {
	flag.IntVar(&port, "port", 6379, "port to listen to")
	log.SetPrefix(fmt.Sprintf("[localhost:%4d] ", port))
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
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	log.Printf("Replication info: %v\n", repl)
	log.Println("Logs from your program will appear here!")

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Printf("Failed to bind to port %d\n", port)
		os.Exit(1)
	}
	defer listener.Close()
	go func() {
		masterConn, remainingResponse, err := repl.HandshakeWithMaster()
		log.Printf("Handshake results: %v, \"%s\", %v\n", masterConn, strings.ReplaceAll(remainingResponse, token.TERMINATOR, "\\r\\n"), err)
		if err != nil {
			log.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if masterConn != nil {
			go handleConnection(masterConn, remainingResponse, true)
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error: ", err.Error())
			continue
		}
		go handleConnection(conn, "", false)
	}
}

func handleConnection(conn net.Conn, startingResponse string, fromMaster bool) {
	log.Printf("%v: %v received connection from %v\n", fromMaster, conn.LocalAddr().String(), conn.RemoteAddr().String())
	connRemoteParts := strings.Split(conn.RemoteAddr().String(), ":")
	connPortString := connRemoteParts[len(connRemoteParts)-1]
	connPort, err := strconv.Atoi(connPortString)
	if err != nil {
		log.Printf("unable to extract port from %s: %v", connPortString, err)
		return
	}
	printCheckMaster := func() {
		log.Printf("port=%4d, fromMaster=%v, isFollower=%v\n", connPort, fromMaster, connPort == repl.MasterPort)
	}
	printCheckMaster()
	if !fromMaster && connPort == repl.MasterPort {
		log.Println("leave connection to be handled by masterConn")
		return
	}
	var data []byte
	for fromMaster == (connPort == repl.MasterPort) {
		printCheckMaster()
		data = make([]byte, 1024)
		// possibly wait for other messages, but fall back on startingResponse if timeout
		if startingResponse != "" {
			conn.SetReadDeadline(time.Now().Add(time.Second))
		}
		dataSize, err := conn.Read(data)
		data = data[:dataSize]
		if err != nil && !(startingResponse != "" && os.IsTimeout(err)) {
			if err == io.EOF {
				log.Println("End of file reached")
			} else {
				log.Println("Error reading from connection: ", err.Error())
			}
			break
		}
		if startingResponse != "" {
			data = []byte(startingResponse + string(data))
			startingResponse = ""
			conn.SetReadDeadline(time.Time{}) // remove timeout
		}
		log.Println("Received: ", strings.ReplaceAll(string(data), token.TERMINATOR, "\\r\\n"))

		// process data
		err = repl.HandshakeWithFollower(conn, data)
		if err == nil {
			continue
		}

		tokens, err := token.ParseInput(string(data))
		var responses []*token.Token
		if err != nil {
			err = fmt.Errorf("error parsing input: %v", err)
			log.Println(err)
			responses = []*token.Token{token.TokeniseError(err)}
		} else {
			log.Println("Tokens:")
			for _, t := range tokens {
				log.Println(*t)
			}
			responses, err = runTokens(tokens)
			if err != nil {
				log.Println(err)
				responses = []*token.Token{token.TokeniseError(err)}
			}
		}
		for _, response := range responses {
			log.Println("Response: ", strings.ReplaceAll(response.EncodedString(), token.TERMINATOR, "\\r\\n"))
			if fromMaster {
				log.Println("Not sending response to master")
			} else {
				_, err = conn.Write([]byte(response.EncodedString()))
				if err != nil {
					log.Println("Error writing to connection: ", err.Error())
					break
				}
			}
		}
	}
	if fromMaster == (connPort == repl.MasterPort) {
		conn.Close()
		// otherwise conn already exists and is being handled by another function
	}
}
