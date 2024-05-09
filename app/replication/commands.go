package replication

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

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
		return &token.Token{Type: token.SimpleStringType, SimpleValue: "0"}, nil
	}
	return nil, nil
}
