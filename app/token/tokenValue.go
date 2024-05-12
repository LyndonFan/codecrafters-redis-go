package token

import (
	"fmt"
	"strconv"
)

const MAX_INT32_NUM_DIGITS = 9

func (tkn *Token) Value() any {
	switch tkn.Type {
	case NullType:
		return nil
	case ErrorType, BulkErrorType:
		return fmt.Errorf(tkn.SimpleValue)
	case SimpleStringType, VerbatimStringType, BulkStringType:
		return tkn.SimpleValue
	case BooleanType:
		switch tkn.SimpleValue {
		case "true":
			return true
		case "false":
			return false
		default:
			panic(fmt.Errorf("expected \"true\" or \"false\", got %s", tkn.SimpleValue))
		}
	case IntegerType:
		x, err := strconv.Atoi(tkn.SimpleValue)
		if err != nil {
			panic(err)
		}
		return x
	case BigNumberType:
		if len(tkn.SimpleValue) < MAX_INT32_NUM_DIGITS || tkn.SimpleValue[0] == '-' {
			x, err := strconv.Atoi(tkn.SimpleValue)
			if err != nil {
				panic(err)
			}
			return x
		}
		var x uint
		for _, c := range tkn.SimpleValue {
			if c == '+' {
				continue
			}
			x = x*10 + uint(c-'0')
		}
		return x
	case ArrayType:
		res := make([]any, len(tkn.NestedValue))
		for i, v := range tkn.NestedValue {
			res[i] = v.Value()
		}
		return res
	case MapType:
		if len(tkn.NestedValue)%2 != 0 {
			panic(fmt.Errorf("expected even number of elements, got %d", len(tkn.NestedValue)))
		}
		res := make(map[any]any)
		for i := 0; i < len(tkn.NestedValue); i += 2 {
			k = tkn.NestedValue[i]
		}
		for k, v := range tkn.NestedValue {

		}
	default:
		panic(fmt.Errorf("unimplemented Token.Value() for type %v", tkn.Type))
	}
}
