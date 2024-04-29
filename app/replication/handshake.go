package replication

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func sendMessage(conn net.Conn, tkn *token.Token) error {
	messageString := tkn.EncodedString()
	fmt.Println("Will send", strings.Replace(messageString, token.TERMINATOR, "\\r\\n", -1))
	_, err = conn.Write([]byte(tkn.EncodedString()))
	if err != nil {
		return err
	}
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	responseString = string(buf[:n])
	if responseString != "+OK"+token.TERMINATOR {
		return fmt.Errorf("received non-OK response: %s", responseString)
	}
	return nil
}

func (r Replicator) HandshakeWithMaster() error {
	if r.IsMaster() {
		fmt.Println("this instance is already the master, will do nothing")
		return nil
	}

	conn, err := net.Dial("tcp", r.MasterAddress())
	if err != nil {
		return err
	}
	defer conn.Close()

	message := &token.Token{
		Type: token.ArrayType,
		NestedValue: []*token.Token{
			{
				Type:        token.BulkStringType,
				SimpleValue: "ping",
			},
		},
	}
	var err error
	err = sendMessage(conn, message)
	if err != nil {
		return err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
		{Type: token.BulkStringType, SimpleValue: "listening-port"},
		{Type: token.BulkStringType, SimpleValue: strconv.Itoa(r.Port)},
	}
	err = sendMessage(conn, message)
	if err != nil {
		return err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
		{Type: token.BulkStringType, SimpleValue: "capa"},
		{Type: token.BulkStringType, SimpleValue: "psync2"},
	}
	err = sendMessage(r.MasterAddress(), message)
	if err != nil {
		return err
	}
	return nil
}
