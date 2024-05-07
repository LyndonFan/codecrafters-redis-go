package replication

import (
	"encoding/hex"
	"fmt"
	"log"
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
			{Type: token.SimpleStringType, SimpleValue: "0"},
		},
	}, nil
}
