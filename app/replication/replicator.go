package replication

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
)

type Replicator struct {
	ID                  string
	Port                int
	MasterHost          string
	MasterPort          int
	MasterRepliID       string
	MasterReplOffset    int
	followerConnections map[int]*net.TCPConn
}

func (r Replicator) String() string {
	if r.IsMaster() {
		return fmt.Sprintf("{master, %s..., %d}", r.MasterRepliID[:6], r.MasterReplOffset)
	}
	return fmt.Sprintf("{%s:%d, %s..., %d}", r.MasterHost, r.MasterPort, r.MasterRepliID[:6], r.MasterReplOffset)
}

func (r Replicator) IsMaster() bool {
	return r.MasterHost == "" && r.MasterPort == 0
}

func (r Replicator) IsFollower(port int) bool {
	_, exists := r.followerConnections[port]
	return exists
}

func (r Replicator) MasterAddress() string {
	if r.IsMaster() {
		return ""
	}
	return fmt.Sprintf("%s:%d", r.MasterHost, r.MasterPort)
}

func (r Replicator) InfoMap() map[string]string {
	role := "slave"
	if r.IsMaster() {
		role = "master"
	}
	return map[string]string{
		"role":               role,
		"master_replid":      r.MasterRepliID,
		"master_repl_offset": strconv.Itoa(r.MasterReplOffset),
	}
}

func GetReplicator(port int, masterHost, masterPortString string) (*Replicator, error) {
	id := randomID()
	connMap := make(map[int]*net.TCPConn)
	if masterHost == "" && masterPortString == "" {
		return &Replicator{ID: id, Port: port, MasterRepliID: id, followerConnections: connMap}, nil
	}
	masterPort, err := strconv.Atoi(masterPortString)
	if err != nil {
		return nil, err
	}
	return &Replicator{
		ID:                  id,
		Port:                port,
		MasterHost:          masterHost,
		MasterPort:          masterPort,
		MasterRepliID:       id,
		MasterReplOffset:    0,
		followerConnections: connMap,
	}, nil
}

const RANDOM_ID_LENGTH int = 40

func randomID() string {
	bytes := make([]byte, RANDOM_ID_LENGTH)
	for i := range bytes {
		x := byte(rand.Intn(16))
		if x < 10 {
			bytes[i] = '0' + x
		} else {
			bytes[i] = 'a' + x - 10
		}
	}
	return string(bytes)
}
