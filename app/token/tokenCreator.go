package token

import (
	"fmt"
	"reflect"
	"strings"
)

func CreateToken(x any) (*Token, error) {
	val := reflect.ValueOf(x)
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	switch {
	case val.Kind() == reflect.Bool:
		return &Token{Type: BooleanType, SimpleValue: fmt.Sprintf("%v", val.Bool())}, nil
	case val.CanFloat():
		return &Token{Type: DoubleType, SimpleValue: fmt.Sprintf("%f", val.Float())}, nil
	case val.CanInt():
		return &Token{Type: IntegerType, SimpleValue: fmt.Sprintf("%d", val.Int())}, nil
	case val.CanUint():
		tknType := IntegerType
		if val.Kind() == reflect.Uint || val.Kind() == reflect.Uint32 || val.Kind() == reflect.Uint64 {
			tknType = BigNumberType
		}
		return &Token{Type: tknType, SimpleValue: fmt.Sprintf("%d", val.Uint())}, nil
	case val.Kind() == reflect.String:
		s := val.String()
		if strings.ContainsAny(s, TERMINATOR) {
			return &Token{Type: BulkStringType, SimpleValue: s}, nil
		}
		return &Token{Type: SimpleStringType, SimpleValue: s}, nil
	case val.Kind() == reflect.Interface: // errors are interfaces in go
		if val.Type().Implements(errorInterface) {
			errorMessage := x.(error).Error()
			tknType := ErrorType
			if strings.ContainsAny(errorMessage, TERMINATOR) {
				tknType = BulkErrorType
			}
			return &Token{Type: tknType, SimpleValue: errorMessage}, nil
		}
	case val.Kind() == reflect.Map:
		mapIter := val.MapRange()
		nestedTokens := make([]*Token, 0, 4)
		idx := 0
		for mapIter.Next() {
			k := mapIter.Key()
			keyKind := reflect.ValueOf(k).Kind()
			if keyKind == reflect.Map || keyKind == reflect.Array || keyKind == reflect.Slice {
				return nil, fmt.Errorf("keys cannot be nested, got %v", keyKind)
			}
			tkn, err := CreateToken(k)
			if err != nil {
				return nil, fmt.Errorf("encountered error for key %d: %v", idx, err)
			}
			nestedTokens = append(nestedTokens, tkn)
			v := mapIter.Value()
			tkn, err = CreateToken(v)
			if err != nil {
				return nil, fmt.Errorf("encountered error for value %d: %v", idx, err)
			}
			nestedTokens = append(nestedTokens, tkn)
			idx++
		}
		return &Token{Type: MapType, NestedValue: nestedTokens}, nil
	case val.Kind() == reflect.Array || val.Kind() == reflect.Slice:
		length := val.Len()
		nestedTokens := make([]*Token, length)
		for i := range nestedTokens {
			v := val.Index(i)
			tkn, err := CreateToken(v)
			if err != nil {
				return nil, fmt.Errorf("encountered error for elem %d: %v", i, err)
			}
			nestedTokens[i] = tkn
		}
		return &Token{Type: ArrayType, NestedValue: nestedTokens}, nil
	case val.IsNil(): // must be checked after array / map
		return &Token{Type: NullType}, nil
	}
	return nil, fmt.Errorf("unable to create token for type %v", reflect.ValueOf(x))
}
