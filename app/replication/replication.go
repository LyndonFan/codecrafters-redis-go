package replication

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

const EMPTY_RDB_FILE_HEX string = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

func (r *Replicator) ShouldAddConnection(port int) bool {
	_, exists := r.followerConnections[port]
	return !exists
}

func (r *Replicator) AddConnection(port int, conn net.Conn) error {
	_, exists := r.followerConnections[port]
	if exists {
		return fmt.Errorf("connection already exists for port %d", port)
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("unable to cast net.Conn to TCPConn")
	}
	tcpConn.SetKeepAlive(true)
	r.followerConnections[port] = tcpConn
	return nil
}

func (r *Replicator) PropagateCommandToken(tkn *token.Token) error {
	return r.PropagateCommandString(tkn.EncodedString())
}

func (r *Replicator) PropagateCommandString(message string) error {
	bytes := []byte(message)
	if len(r.followerConnections) > 0 {
		fmt.Printf("Will replicate this command to %d ports: %s\n", len(r.followerConnections), replaceTerminator(message))
		fmt.Println(r.followerConnections)
	} else {
		fmt.Println("No followers to replicate to")
	}
	for port, conn := range r.followerConnections {
		fmt.Println("Replicating to port", port)
		n, err := conn.Write(bytes)
		fmt.Printf("Sent %d bytes, ", n)
		if err != nil {
			fmt.Println("error: ", err.Error())
			return err
		}
		fmt.Println("success")
	}
	return nil
}
