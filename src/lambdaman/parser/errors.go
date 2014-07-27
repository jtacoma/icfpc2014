package parser

import "fmt"

type UnsupportedTypeError struct {
	Instance interface{}
}

func (e UnsupportedTypeError) Error() string {
	return fmt.Sprintf("unsupported type %T", e.Instance)
}
