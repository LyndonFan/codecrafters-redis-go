package replication

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

const EMPTY_RDB_FILE_HEX string = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

func (r *Replicator) AddFollower(port int) {
	r.FollowerPorts = append(r.FollowerPorts, port)
}

func (r *Replicator) GetFollowerConnection(port int) (*net.TCPConn, error) {
	if conn, exists := r.followerConnections[port]; exists {
		return conn, nil
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("connection isn't TCP")
	}
	r.followerConnections[port] = tcpConn
	return tcpConn, nil
}

func (r *Replicator) PropagateCommand(tkn *token.Token) error {
	bytes := []byte(tkn.EncodedString())
	for _, port := range r.FollowerPorts {
		conn, err := r.GetFollowerConnection(port)
		if err != nil {
			return err
		}
		_, err = conn.Write(bytes)
		if err != nil {
			return err
		}
	}
	return nil
}
