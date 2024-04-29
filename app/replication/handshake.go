package replication

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func sendMessage(conn net.Conn, tkn *token.Token) (string, error) {
	messageString := tkn.EncodedString()
	fmt.Println("Will send", replaceTerminator(messageString))
	_, err := conn.Write([]byte(tkn.EncodedString()))
	if err != nil {
		return "", err
	}
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	response := string(buf[:n])
	return response, nil
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

	var response string
	response, err = sendMessage(conn, message)
	if err != nil {
		return err
	}
	err = checkString(response, "+PONG"+token.TERMINATOR)
	if err != nil {
		return err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
		{Type: token.BulkStringType, SimpleValue: "listening-port"},
		{Type: token.BulkStringType, SimpleValue: strconv.Itoa(r.Port)},
	}
	response, err = sendMessage(conn, message)
	if err != nil {
		return err
	}
	err = checkString(response, token.OKToken.EncodedString())
	if err != nil {
		return err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
		{Type: token.BulkStringType, SimpleValue: "capa"},
		{Type: token.BulkStringType, SimpleValue: "psync2"},
	}
	response, err = sendMessage(conn, message)
	if err != nil {
		return err
	}
	err = checkString(response, token.OKToken.EncodedString())
	if err != nil {
		return err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "PSYNC"},
		{Type: token.BulkStringType, SimpleValue: "?"},
		{Type: token.BulkStringType, SimpleValue: "-1"},
	}
	response, err = sendMessage(conn, message)
	if err != nil {
		return err
	}
	expectedStringPattern := "\\+FULLRESYNC [a-z0-9]{40} 0\r\n" // need double escape
	if matched, err := regexp.MatchString(expectedStringPattern, response); !matched || err != nil {
		return fmt.Errorf("expected response to match \"%s\", got %s", replaceTerminator(expectedStringPattern), replaceTerminator(response))
	}
	return nil
}

func replaceTerminator(x string) string {
	return strings.Replace(x, token.TERMINATOR, "\\r\\n", -1)
}

func checkString(actual, expected string) error {
	if actual == expected {
		return nil
	}
	return fmt.Errorf(
		"expected \"%s\", received \"%s\"",
		replaceTerminator(expected),
		replaceTerminator(actual),
	)
}
