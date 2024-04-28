package replication

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func (r Replicator) HandshakeWithMaster() error {
	if r.IsMaster() {
		return fmt.Errorf("this instance is already the master")
	}
	conn, err := net.Dial("tcp", r.MasterAddress())
	if err != nil {
		return err
	}
	message := token.Token{
		Type: token.ArrayType,
		NestedValue: []*token.Token{
			&token.Token{
				Type:        token.BulkStringType,
				SimpleValue: "ping",
			},
		},
	}
	messageString := message.EncodedString()
	fmt.Println("Will send", strings.Replace(messageString, token.TERMINATOR, "\\r\\n", -1))
	_, err = conn.Write([]byte(message.EncodedString()))
	return err
}
