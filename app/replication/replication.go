package replication

import (
	b64 "encoding/base64"
	"fmt"
	"net"
)

const EMPTY_RDB_FILE_BASE64 string = "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="

func (r *Replicator) AddFollower(port int) {
	r.FollowerPorts = append(r.FollowerPorts, port)
}

func (r *Replicator) SendRDBFileToFollowers() error {
	decoded, err := b64.RawStdEncoding.DecodeString(EMPTY_RDB_FILE_BASE64)
	if err != nil {
		return err
	}
	for _, fPort := range r.FollowerPorts {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", fPort))
		if err != nil {
			return err
		}
		bytes := fmt.Sprintf("$%d\r\n%s", len(decoded), string(decoded))
		_, err = conn.Write([]byte(bytes))
		conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
