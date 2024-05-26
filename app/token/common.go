package token

import (
	"fmt"
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

// same as `TokeniseError(fmt.Errorf(s, a...))`
func TokeniseErrorf(s string, a ...any) *Token {
	errString := fmt.Sprintf(s, a...)
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
