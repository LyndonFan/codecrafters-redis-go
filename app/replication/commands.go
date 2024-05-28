package replication

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func (repl *Replicator) RespondToPsync(ctx context.Context, args []any) (*token.Token, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	returnValue := fmt.Sprintf("FULLRESYNC %s %d%s", repl.MasterReplID, repl.MasterReplOffset, token.TERMINATOR)
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

func (repl *Replicator) RespondToReplconf(ctx context.Context, args []any) (*token.Token, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	stringArgs := make([]string, len(args))
	for i, v := range args {
		stringArgs[i] = strings.ToLower(fmt.Sprintf("%v", v))
	}
	repl.logger.Debug(fmt.Sprintf("stringArgs %v", stringArgs))
	if stringArgs[0] == "listening-port" {
		return nil, fmt.Errorf("this should be handled separately")
	}
	if stringArgs[0] == "capa" && stringArgs[1] == "psync2" {
		return &token.OKToken, nil
	}
	if stringArgs[0] == "getack" && stringArgs[1] == "*" {
		return &token.Token{
			Type: token.ArrayType,
			NestedValue: []*token.Token{
				{Type: token.SimpleStringType, SimpleValue: "REPLCONF"},
				{Type: token.SimpleStringType, SimpleValue: "ACK"},
				{Type: token.SimpleStringType, SimpleValue: strconv.Itoa(repl.BytesProcessed)},
			},
		}, nil
	}
	if stringArgs[0] == "ack" {
		followerAddress, ok := ctx.Value("address").(string)
		if !ok {
			return nil, fmt.Errorf("unable to find address from context")
		}
		followerAddressSplit := strings.Split(followerAddress, ":")
		followerPort, err := strconv.Atoi(followerAddressSplit[len(followerAddressSplit)-1])
		if err != nil {
			return nil, fmt.Errorf("unable to get port from context: %v", err)
		}
		err = repl.followerCounter.AddRespondedFollower(followerPort)
		if err == ErrNotLocked {
			err = nil
		}
		return nil, err
	}
	return nil, fmt.Errorf("not implemented yet")
}

func (repl *Replicator) RespondToWait(ctx context.Context, args []any) (*token.Token, error) {
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
	repl.logger.Debugf("Start getting ACKs from followers")
	res, err := repl.countAckFromFollowers(numReplicas, timeout)
	repl.logger.Debugf("Finished count ACKs, got %d with errs %v", res, err)
	if err != nil {
		return nil, err
	}
	return &token.Token{Type: token.IntegerType, SimpleValue: strconv.Itoa(res)}, nil
}

func (repl *Replicator) countAckFromFollowers(numReplicas, timeoutMilliseconds int) (int, error) {
	repl.followerCounter.StartBlock()
	defer repl.followerCounter.EndBlock()
	askToken := &token.Token{
		Type: token.ArrayType,
		NestedValue: []*token.Token{
			{Type: token.BulkStringType, SimpleValue: "REPLCONF"},
			{Type: token.BulkStringType, SimpleValue: "GETACK"},
			{Type: token.BulkStringType, SimpleValue: "*"},
		},
	}
	message := askToken.EncodedString()
	errs := make(map[int]error)

	doneChannel := make(chan bool)
	defer close(doneChannel)
	if timeoutMilliseconds == 0 {
		go func() {
			for repl.followerCounter.Locked() {
				port, ok := <-repl.followerCounter.portChannel
				if !ok {
					break
				}
				repl.logger.Debug(fmt.Sprintf("Received ack from port %v\n", port))
				count := repl.followerCounter.NumRespondedFollowers()
				if count >= numReplicas || count == len(repl.followerConnections) {
					doneChannel <- true
					break
				}
			}
		}()
	}
	respond := func(port int, conn *net.TCPConn) {
		repl.logger.Debug(fmt.Sprintf("[follower:%04d] asking ACK from follower\n", port))
		_, err := conn.Write([]byte(message))
		if err != nil {
			repl.logger.Debug(fmt.Sprintf("[follower:%04d] encountered error from follower: %v\n", port, err))
			errs[port] = err
		}
	}
	repl.logger.Debug(fmt.Sprintf("Start sending \"REPLCONF GETACK *\" to %v follower(s)\n", len(repl.followerConnections)))
	for port, conn := range repl.followerConnections {
		go respond(port, conn)
	}
	if timeoutMilliseconds == 0 {
		repl.logger.Debug("No timeout specified, so waiting to hear from enough replicas")
		<-doneChannel
		repl.logger.Debug("Heard back from sufficient (or all) replicas")
	} else {
		repl.logger.Debugf("Timeout specified, so sleeping for %dms", timeoutMilliseconds)
		time.Sleep(time.Millisecond * time.Duration(timeoutMilliseconds))
		repl.logger.Debug("Timeout ended")
	}
	repl.logger.Debugf("errors: %v", errs)
	if len(errs) == 0 {
		return repl.followerCounter.NumRespondedFollowers(), nil
	}
	errorMessage := "Encountered these errors when waiting for followers:"
	for port, e := range errs {
		errorMessage += fmt.Sprintf("\n%04d: %v", port, e)
	}
	return 0, fmt.Errorf(errorMessage)
}
