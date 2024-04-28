package replication

import (
	"fmt"
	"math/rand"
	"strconv"
)

type ReplicationInfo struct {
	Host             string
	Port             int
	MasterRepliID    string
	MasterReplOffset int
}

func (ri ReplicationInfo) String() string {
	if ri.IsMaster() {
		return fmt.Sprintf("{master, %s..., %d}", ri.MasterRepliID[:6], ri.MasterReplOffset)
	}
	return fmt.Sprintf("{%s:%d, %s..., %d}", ri.Host, ri.Port, ri.MasterRepliID[:6], ri.MasterReplOffset)
}

func (ri ReplicationInfo) IsMaster() bool {
	return ri.Host == "" && ri.Port == 0
}

func (ri ReplicationInfo) InfoMap() map[string]string {
	role := "slave"
	if ri.IsMaster() {
		role = "master"
	}
	return map[string]string{
		"role":               role,
		"master_replid":      ri.MasterRepliID,
		"master_repl_offset": strconv.Itoa(ri.MasterReplOffset),
	}
}

func GetReplicationInfo(host, portString string) (*ReplicationInfo, error) {
	id := randomID()
	if host == "" && portString == "" {
		return &ReplicationInfo{MasterRepliID: id}, nil
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, err
	}
	return &ReplicationInfo{
		Host:             host,
		Port:             port,
		MasterRepliID:    id,
		MasterReplOffset: 0,
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
