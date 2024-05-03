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
	fmt.Println("Success")
	return nil
}

const REPLCONF_LISTENING_PORT_PREFIX = "*3\r\n$8\r\nreplconf\r\n$14\r\nlistening-port\r\n$"

func (r Replicator) HandshakeWithFollower(conn net.Conn, message []byte) error {
	messageString := string(message)
	if len(messageString) < len(REPLCONF_LISTENING_PORT_PREFIX) {
		return fmt.Errorf("expected at least %d bytes, got %d", len(REPLCONF_LISTENING_PORT_PREFIX), len(messageString))
	}
	messagePrefix := messageString[:len(REPLCONF_LISTENING_PORT_PREFIX)]
	if strings.ToLower(messagePrefix) != REPLCONF_LISTENING_PORT_PREFIX {
		return fmt.Errorf("expected prefix %s, got %s", replaceTerminator(REPLCONF_LISTENING_PORT_PREFIX), replaceTerminator(messageString))
	}
	tokens, err := token.ParseInput(messageString)
	if err != nil {
		return err
	}
	if len(tokens) != 1 {
		return fmt.Errorf("expected 1 token, got %d", len(tokens))
	}
	tkn := tokens[0]
	if tkn.Type != token.ArrayType {
		return fmt.Errorf("expected array token, got %s", tkn.Type)
	}
	if len(tkn.NestedValue) != 3 {
		return fmt.Errorf("expected 3 tokens, got %d", len(tkn.NestedValue))
	}
	port, err := strconv.Atoi(tkn.NestedValue[2].SimpleValue)
	if err != nil {
		return err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("unable to cast net.Conn to TCPConn")
	}
	tcpConn.SetKeepAlive(true)
	r.followerConnections[port] = tcpConn
	fmt.Println("Handshake with follower succeeded in 2nd step, saved connection")

	response := token.OKToken.EncodedString()
	_, err = conn.Write([]byte(response))
	return err
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
