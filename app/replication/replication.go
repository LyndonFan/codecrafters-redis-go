package replication

import (
	"fmt"
	"math/rand"
	"strconv"
)

type Replicator struct {
	Host             string
	Port             int
	MasterRepliID    string
	MasterReplOffset int
}

func (r Replicator) String() string {
	if r.IsMaster() {
		return fmt.Sprintf("{master, %s..., %d}", r.MasterRepliID[:6], r.MasterReplOffset)
	}
	return fmt.Sprintf("{%s:%d, %s..., %d}", r.Host, r.Port, r.MasterRepliID[:6], r.MasterReplOffset)
}

func (r Replicator) IsMaster() bool {
	return r.Host == "" && r.Port == 0
}

func (r Replicator) MasterAddress() string {
	if !r.IsMaster() {
		return fmt.Sprintf("%s:%d", r.Host, r.Port)
	}
	return ""
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

func GetReplicator(host, portString string) (*Replicator, error) {
	id := randomID()
	if host == "" && portString == "" {
		return &Replicator{MasterRepliID: id}, nil
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, err
	}
	return &Replicator{
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
