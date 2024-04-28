package token

var NullBulkString Token = Token{Type: BulkStringType, SimpleValue: "", representNull: true}

var OKToken Token = Token{Type: SimpleStringType, SimpleValue: "OK"}
