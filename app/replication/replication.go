package replication

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

const EMPTY_RDB_FILE_HEX string = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

func (repl *Replicator) ShouldAddConnection(port int) bool {
	_, exists := repl.followerConnections[port]
	return !exists
}

func (repl *Replicator) AddConnection(port int, conn net.Conn) error {
	_, exists := repl.followerConnections[port]
	if exists {
		return fmt.Errorf("connection already exists for port %d", port)
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("unable to cast net.Conn to TCPConn")
	}
	tcpConn.SetKeepAlive(true)
	repl.followerConnections[port] = tcpConn
	return nil
}

func (repl *Replicator) PropagateCommandToken(tkn *token.Token) error {
	return repl.PropagateCommandString(tkn.EncodedString())
}

func (repl *Replicator) PropagateCommandString(message string) error {
	bytes := []byte(message)
	if len(repl.followerConnections) > 0 {
		repl.logger.Debug(fmt.Sprintf("Will replicate this command to %d ports: %s\n", len(repl.followerConnections), replaceTerminator(message)))
		repl.logger.Debug(fmt.Sprintf("%v", repl.followerConnections))
	} else {
		repl.logger.Debug("No followers to replicate to")
	}
	for port, conn := range repl.followerConnections {
		repl.logger.Debug(fmt.Sprintf("Replicating to port %v", port))
		n, err := conn.Write(bytes)
		repl.logger.Debug(fmt.Sprintf("Sent %d bytes, ", n))
		if err != nil {
			repl.logger.Debug(fmt.Sprintf("error:  %v", err.Error()))
			return err
		}
		repl.logger.Debug("success")
	}
	return nil
}
