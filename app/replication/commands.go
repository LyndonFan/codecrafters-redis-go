package replication

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func (repl *Replicator) RespondToPsync(args []any) (*token.Token, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	returnValue := fmt.Sprintf("FULLRESYNC %s %d%s", repl.MasterRepliID, repl.MasterReplOffset, token.TERMINATOR)
	emptyRDB, err := hex.DecodeString(EMPTY_RDB_FILE_HEX)
	if err != nil {
		return nil, err
	}
	returnValue += fmt.Sprintf("$%d\r\n%s", len(emptyRDB), string(emptyRDB))
	tkn := token.Token{
		Type:        token.SimpleStringType,
		SimpleValue: returnValue,
	}
	tkn.StripTrailingTerminator()
	return &tkn, nil
}

func (repl *Replicator) RespondToReplconf(args []any) (*token.Token, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	stringArgs := make([]string, len(args))
	for i, v := range args {
		stringArgs[i] = strings.ToLower(fmt.Sprintf("%v", v))
	}
	log.Println("stringArgs", stringArgs)
	if stringArgs[0] == "listening-port" {
		return nil, fmt.Errorf("this should be handled separately")
	}
	if stringArgs[0] != "getack" {
		return &token.OKToken, nil
	}
	if stringArgs[1] != "*" {
		return nil, fmt.Errorf("not implemented yet")
	}
	return &token.Token{
		Type: token.ArrayType,
		NestedValue: []*token.Token{
			{Type: token.SimpleStringType, SimpleValue: "REPLCONF"},
			{Type: token.SimpleStringType, SimpleValue: "ACK"},
			{Type: token.SimpleStringType, SimpleValue: strconv.Itoa(repl.BytesProcessed)},
		},
	}, nil
}

func (repl *Replicator) RespondToWait(args []any) (*token.Token, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	arg0, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("expected arguments to be strings, got %v", args[0])
	}
	arg1, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("expected arguments to be strings, got %v", args[1])
	}
	numReplicas, err := strconv.Atoi(arg0)
	if err != nil || numReplicas < 0 {
		return nil, fmt.Errorf("expected first argument to be like a positive number, got %s", arg0)
	}
	timeout, err := strconv.Atoi(arg1)
	if err != nil || timeout < 0 {
		return nil, fmt.Errorf("expected first argument to be like a positive number, got %s", arg1)
	}
	if numReplicas == 0 {
		return &token.Token{Type: token.IntegerType, SimpleValue: "0"}, nil
	}
	success := repl.startBlock()
	if !success {
		return nil, fmt.Errorf("unable to start block for replicator")
	}
	defer repl.endBlock()
	res, err := repl.countAckFromFollowers(numReplicas, timeout)
	if err != nil {
		return nil, err
	}
	return &token.Token{Type: token.IntegerType, SimpleValue: strconv.Itoa(res)}, nil
}

func (repl *Replicator) countAckFromFollowers(numReplicas, timeoutSeconds int) (int, error) {
	askToken := &token.Token{
		Type: token.ArrayType,
		NestedValue: []*token.Token{
			{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
			{Type: token.BulkStringType, SimpleValue: "GETACK"},
			{Type: token.BulkStringType, SimpleValue: "*"},
		},
	}
	message := askToken.EncodedString()
	count := 0
	errs := make(map[int]error)
	addChannel, doneChannel := make(chan int), make(chan bool)
	checkDone := func(add chan int, done chan bool) {
		defer close(add)
		var ok bool
		for {
			_, ok = <-add
			if !ok {
				break
			}
			count++
			if count >= numReplicas {
				done <- true
				break
			}
		}
	}
	go checkDone(addChannel, doneChannel)
	respond := func(port int, conn *net.TCPConn) {
		err := repl.getAckFromFollower(conn, timeoutSeconds, message, addChannel)
		if err != nil {
			errs[port] = err
		}
	}
	for port, conn := range repl.followerConnections {
		go respond(port, conn)
	}
	select {
	case <-doneChannel:
		log.Println("Heard back from sufficient replicas before timeout")
	case <-time.After(time.Second * time.Duration(timeoutSeconds)):
		log.Println("Haven't got enough replies but shutting early due to timeout")
	}
	if len(errs) == 0 {
		return count, nil
	}
	errorMessage := "Encountered these errors when waiting for followers:"
	for port, e := range errs {
		errorMessage += fmt.Sprintf("\n%04d: %v", port, e)
	}
	return 0, fmt.Errorf(errorMessage)
}

func (repl *Replicator) getAckFromFollower(conn *net.TCPConn, timeoutSeconds int, message string, addChannel chan int) error {
	defer conn.SetReadDeadline(time.Time{})
	data := []byte(message)
	_, err := conn.Write(data)
	if err != nil {
		return err
	}
	errCh := make(chan error)
	go func() {
		response := make([]byte, 1024)
		n, err := conn.Read(response)
		response = response[:n]
		if err != nil {
			errCh <- err
			return
		}
		resTokens, err := token.ParseInput(string(response))
		if err != nil {
			errCh <- err
			return
		}
		if len(resTokens) != 1 {
			errCh <- fmt.Errorf("expected 1 token, got %d", len(resTokens))
			return
		}
		x, isInt := (resTokens[0].Value()).(int)
		if !isInt {
			errCh <- fmt.Errorf("returned response isn't an integer")
			return
		}
		errCh <- nil
		addChannel <- x
	}()
	select {
	case resErr := <-errCh:
		return resErr
	case <-time.After(time.Second * time.Duration(timeoutSeconds)):
		return fmt.Errorf("TIMED OUT")
	}
}
