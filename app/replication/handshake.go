package replication

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func sendMessage(conn net.Conn, tkn *token.Token, expectedResponse string) error {
	messageString := tkn.EncodedString()
	fmt.Println("Will send", strings.Replace(messageString, token.TERMINATOR, "\\r\\n", -1))
	_, err := conn.Write([]byte(tkn.EncodedString()))
	if err != nil {
		return err
	}
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	response := string(buf[:n])
	if response != expectedResponse {
		return fmt.Errorf(
			"expected \"%s\", received %s", strings.Replace(expectedResponse, token.TERMINATOR, "\\r\\n", -1), strings.Replace(response, token.TERMINATOR, "\\r\\n", -1))
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
	err = sendMessage(conn, message, "+PONG"+token.TERMINATOR)
	if err != nil {
		return err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
		{Type: token.BulkStringType, SimpleValue: "listening-port"},
		{Type: token.BulkStringType, SimpleValue: strconv.Itoa(r.Port)},
	}
	err = sendMessage(conn, message, token.OKToken.EncodedString())
	if err != nil {
		return err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
		{Type: token.BulkStringType, SimpleValue: "capa"},
		{Type: token.BulkStringType, SimpleValue: "psync2"},
	}
	err = sendMessage(conn, message, token.OKToken.EncodedString())
	if err != nil {
		return err
	}
	return nil
}
