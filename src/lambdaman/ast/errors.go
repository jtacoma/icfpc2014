package ast

import (
	"bytes"
	"fmt"
)

type NotFoundError struct {
	frame     *Frame
	BlockName string
	VarName   string
}

func (e NotFoundError) Error() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%v not available in %s",
		e.VarName, e.BlockName)
	iframe := 0
	for frame := e.frame; frame != nil; frame = frame.Parent {
		fmt.Fprintf(&buf, "\n  frame %d:", iframe)
		for _, datum := range frame.Data {
			fmt.Fprintf(&buf, " %s", datum.Name)
		}
		if len(frame.Data) == 0 {
			fmt.Fprint(&buf, " (empty)")
		}
		iframe += 1
	}
	return buf.String()
}
