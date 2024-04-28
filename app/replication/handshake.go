package replication

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func sendMessage(address string, tkn *token.Token) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()
	messageString := tkn.EncodedString()
	fmt.Println("Will send", strings.Replace(messageString, token.TERMINATOR, "\\r\\n", -1))
	_, err = conn.Write([]byte(tkn.EncodedString()))
	return err
}

func (r Replicator) HandshakeWithMaster() error {
	if r.IsMaster() {
		fmt.Println("this instance is already the master, will do nothing")
		return nil
	}

	message := &token.Token{
		Type: token.ArrayType,
		NestedValue: []*token.Token{
			&token.Token{
				Type:        token.BulkStringType,
				SimpleValue: "ping",
			},
		},
	}
	var err error
	err = sendMessage(r.MasterAddress(), message)
	if err != nil {
		return err
	}

	message.NestedValue = []*token.Token{
		&token.Token{
			Type:        token.BulkStringType,
			SimpleValue: "REPLCONF",
		},
		&token.Token{
			Type:        token.BulkStringType,
			SimpleValue: "listening-port",
		},
		&token.Token{
			Type:        token.BulkStringType,
			SimpleValue: strconv.Itoa(r.Port),
		},
	}
	err = sendMessage(r.MasterAddress(), message)
	if err != nil {
		return err
	}

	message.NestedValue = []*token.Token{
		&token.Token{
			Type:        token.BulkStringType,
			SimpleValue: "REPLCONF",
		},
		&token.Token{
			Type:        token.BulkStringType,
			SimpleValue: "capa",
		},
		&token.Token{
			Type:        token.BulkStringType,
			SimpleValue: "psync2",
		},
	}
	err = sendMessage(r.MasterAddress(), message)
	if err != nil {
		return err
	}
	return nil
}
