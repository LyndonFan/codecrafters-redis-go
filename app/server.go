package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	customLogger "github.com/codecrafters-io/redis-starter-go/app/logger"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/token"
)

var port int
var logLevel string
var logger *slog.Logger
var repl *replication.Replicator

func init() {
	flag.IntVar(&port, "port", 6379, "port to listen to")
	log.SetPrefix(fmt.Sprintf("[localhost:%4d] ", port))
	flag.StringVar(&logLevel, "loglevel", customLogger.LOG_LEVEL_INFO, "log level of messages")
	logger = customLogger.NewLogger(port, logLevel)
	var replHost string
	flag.StringVar(&replHost, "replicaof", "", "if specified, the host and port of its master. Works with `--replicaof HOST PORT` or `--replicaof \"HOST PORT\"`")
	flag.Parse()
	remainingArgs := flag.Args()
	var err error
	replPortString := "" // not 0, as it could be a possible valid value
	if len(remainingArgs) > 0 {
		replPortString = remainingArgs[0]
	} else if strings.Contains(replHost, " ") {
		parts := strings.Split(replHost, " ")
		replHost, replPortString = parts[0], parts[1]
	}
	if replHost != "" && replPortString != "" {
		repl, err = replication.GetReplicator(port, replHost, replPortString)
	} else {
		repl, err = replication.GetReplicator(port, "", "")
	}
	if err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
}

func main() {
	logger.Debug(fmt.Sprintf("Replication info: %v\n", repl))

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to bind to port %d\n", port))
		os.Exit(1)
	}
	defer listener.Close()
	mainContext := context.Background()
	go func() {
		masterConn, remainingResponse, err := repl.HandshakeWithMaster()
		logger.Debug(fmt.Sprintf("Handshake results: %v, \"%s\", %v\n", masterConn, strings.ReplaceAll(remainingResponse, token.TERMINATOR, "\\r\\n"), err))
		if err != nil {
			logger.Error(fmt.Sprintf("Error: %v", err))
			os.Exit(1)
		}
		if masterConn != nil {
			ctx := context.WithValue(mainContext, "fromMaster", true)
			ctx = context.WithValue(ctx, "address", masterConn.RemoteAddr().String())
			go handleConnection(ctx, masterConn, remainingResponse, true)
		}
	}()
	for {
		if repl.Blocked() {
			time.Sleep(time.Second)
			continue
		}
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(fmt.Sprintf("Error:  %v", err.Error()))
			continue
		}
		ctx := context.WithValue(mainContext, "fromMaster", true)
		ctx = context.WithValue(ctx, "address", conn.RemoteAddr().String())
		go handleConnection(ctx, conn, "", false)
	}
}

func handleConnection(ctx context.Context, conn net.Conn, startingResponse string, fromMaster bool) {
	logger.Debug(fmt.Sprintf("%v: %v received connection from %v\n", fromMaster, conn.LocalAddr().String(), conn.RemoteAddr().String()))
	connRemoteParts := strings.Split(conn.RemoteAddr().String(), ":")
	connPortString := connRemoteParts[len(connRemoteParts)-1]
	connPort, err := strconv.Atoi(connPortString)
	if err != nil {
		logger.Error(fmt.Sprintf("unable to extract port from %s: %v", connPortString, err))
		return
	}
	printCheckMaster := func() {
		logger.Debug(fmt.Sprintf("port=%4d, fromMaster=%v, isFollower=%v\n", connPort, fromMaster, connPort == repl.MasterPort))
	}
	printCheckMaster()
	if !fromMaster && connPort == repl.MasterPort {
		logger.Debug("leave connection to be handled by masterConn")
		return
	}
	var data []byte
	for fromMaster == (connPort == repl.MasterPort) {
		if repl.Blocked() {
			time.Sleep(time.Second)
			continue
		}
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
				logger.Warn("End of file reached")
			} else {
				logger.Error(fmt.Sprintf("Error reading from connection:  %v", err.Error()))
			}
			break
		}
		if startingResponse != "" {
			data = []byte(startingResponse + string(data))
			startingResponse = ""
			conn.SetReadDeadline(time.Time{}) // remove timeout
		}
		logger.Debug(fmt.Sprintf("Received:  %v", strings.ReplaceAll(string(data)), token.TERMINATOR, "\\r\\n"))

		// process data
		err = repl.HandshakeWithFollower(conn, data)
		if err == nil {
			continue
		}

		tokens, err := token.ParseInput(string(data))
		var responses []*token.Token
		if err != nil {
			logger.Error(fmt.Sprintf("error parsing input: %v", err))
			responses = []*token.Token{token.TokeniseError(err)}
		} else {
			log.Println("Tokens:")
			for _, t := range tokens {
				log.Println(*t)
			}
			responses, err = runTokens(ctx, tokens)
			if err != nil {
				logger.Error(err.Error())
				responses = []*token.Token{token.TokeniseError(err)}
			}
		}
		for _, response := range responses {
			logger.Debug(fmt.Sprintf("Response:  %v", strings.ReplaceAll(response.EncodedString()), token.TERMINATOR, "\\r\\n"))
			if response.EncodedString() == "" {
				logger.Warn("Nothing to write, skipping")
				continue
			}
			// TODO: patch this workaround?
			if fromMaster && !strings.Contains(response.EncodedString(), "ACK\r\n") {
				logger.Warn("Not sending response to master")
			} else {
				_, err = conn.Write([]byte(response.EncodedString()))
				if err != nil {
					logger.Error(fmt.Sprintf("Error writing to connection:  %v", err.Error()))
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
