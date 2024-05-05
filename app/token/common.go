package token

import (
	"strings"
)

func TokeniseError(err error) *Token {
	errString := err.Error()
	errType := ErrorType
	if strings.ContainsAny(errString, "\r\n") {
		errType = BulkErrorType
	}
	return &Token{Type: errType, SimpleValue: errString}
}

// func (tkn *Token) String() string {
// 	if tkn.NestedValue == nil {
// 		return fmt.Sprintf("{Type: %s, Value: %s}", tkn.Type, tkn.Nest)
// 	}
// }
