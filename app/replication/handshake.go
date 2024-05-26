package replication

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

var HANDSHAKE_FINAL_PATTERN *regexp.Regexp = regexp.MustCompile("^\\+FULLRESYNC [a-z0-9]{40} \\d+\r\n")

func (repl Replicator) HandshakeWithMaster() (net.Conn, string, error) {
	if repl.IsMaster() {
		repl.logger.Debug("this instance is already the master, will do nothing")
		return nil, "", nil
	}

	conn, err := net.Dial("tcp", repl.MasterAddress())
	if err != nil {
		return nil, "", err
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
	response, err = repl.sendMessage(conn, message)
	if err != nil {
		return nil, "", err
	}
	err = checkString(response, "+PONG"+token.TERMINATOR)
	if err != nil {
		return nil, "", err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
		{Type: token.BulkStringType, SimpleValue: "listening-port"},
		{Type: token.BulkStringType, SimpleValue: strconv.Itoa(repl.Port)},
	}
	response, err = repl.sendMessage(conn, message)
	if err != nil {
		return nil, "", err
	}
	err = checkString(response, token.OKToken.EncodedString())
	if err != nil {
		return nil, "", err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
		{Type: token.BulkStringType, SimpleValue: "capa"},
		{Type: token.BulkStringType, SimpleValue: "psync2"},
	}
	response, err = repl.sendMessage(conn, message)
	if err != nil {
		return nil, "", err
	}
	err = checkString(response, token.OKToken.EncodedString())
	if err != nil {
		return nil, "", err
	}

	message.NestedValue = []*token.Token{
		{Type: token.BulkStringType, SimpleValue: "PSYNC"},
		{Type: token.BulkStringType, SimpleValue: "?"},
		{Type: token.BulkStringType, SimpleValue: "-1"},
	}
	response, err = repl.sendMessage(conn, message)
	if err != nil {
		return nil, "", err
	}
	repl.logger.Debug(fmt.Sprintf("Handshake final message: %v", replaceTerminator(response)))
	matches := HANDSHAKE_FINAL_PATTERN.FindStringSubmatch(response)
	if len(matches) == 0 {
		return nil, "", fmt.Errorf("expected response to match \"%s\", got %s", replaceTerminator(HANDSHAKE_FINAL_PATTERN.String()), replaceTerminator(response))
	}
	if len(matches) > 1 {
		return nil, "", fmt.Errorf("expected exactly 1 match, got %d", len(matches))
	}
	remainingResponse := response[len(matches[0]):]
	remainingResponse, err = repl.readRestRDB(conn, remainingResponse)
	if err != nil {
		return nil, "", err
	}
	repl.logger.Debug("Success")
	return conn, remainingResponse, nil
}

const REPLCONF_LISTENING_PORT_PREFIX = "*3\r\n$8\r\nreplconf\r\n$14\r\nlistening-port\r\n$"

func (repl Replicator) HandshakeWithFollower(conn net.Conn, message []byte) error {
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
	repl.followerConnections[port] = tcpConn
	repl.logger.Debug("Handshake with follower succeeded in 2nd step, saved connection")

	response := token.OKToken.EncodedString()
	_, err = conn.Write([]byte(response))
	return err
}

var RDB_HEADER_PATTERN = regexp.MustCompile("^\\$\\d+\r\n")

func (repl *Replicator) readRestRDB(conn net.Conn, remainingResponse string) (string, error) {
	// expected RDB file response: $<length>\r\n[.]{<length>}
	matches := RDB_HEADER_PATTERN.FindStringSubmatch(remainingResponse)
	var data []byte
	for len(matches) == 0 {
		data = make([]byte, 128) // intentionally small
		n, err := conn.Read(data)
		if err != nil {
			return "", err
		}
		remainingResponse += string(data[:n])
		matches = RDB_HEADER_PATTERN.FindStringSubmatch(remainingResponse)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("expected exactly 1 match for header when reading RDB, got %d", len(matches))
	}
	rdbLengthString := strings.Trim(matches[0], "$\r\n")
	rdbLength, err := strconv.Atoi(rdbLengthString)
	if err != nil {
		return "", err
	}
	remainingResponse = remainingResponse[len(matches[0]):]
	for len(remainingResponse) < rdbLength {
		data = make([]byte, 128) // intentionally small
		n, err := conn.Read(data)
		if err != nil {
			return "", err
		}
		remainingResponse += string(data[:n])
	}
	repl.logger.Debug(fmt.Sprintf("Received RDB file: %s", replaceTerminator(matches[0]+remainingResponse[:rdbLength])))
	remainingResponse = remainingResponse[rdbLength:]
	return remainingResponse, nil
}

func (repl *Replicator) sendMessage(conn net.Conn, tkn *token.Token) (string, error) {
	messageString := tkn.EncodedString()
	repl.logger.Debug(fmt.Sprintf("Will send %v", replaceTerminator(messageString)))
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
