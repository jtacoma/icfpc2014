package ast

import (
	"errors"
	"fmt"
	"io"
)

type Program struct {
	Name      string
	functions []*Function
}

func (p *Program) NewFunction(name string) *Function {
	function := &Function{
		Name: name,
	}
	p.functions = append(p.functions, function)
	return function
}

func (p *Program) WriteTo(w io.Writer) (err error) {
	if _, err = fmt.Fprintln(w, "; program:", p.Name); err != nil {
		return
	}
	var (
		preamble Commands
		offset   = len(p.functions) + 2
		main     *Function
		nonmain  []*Function
	)
	for i := range p.functions {
		if p.functions[i].Name == "main" {
			main = p.functions[i]
		} else {
			nonmain = append(nonmain, p.functions[i])
		}
	}
	if main == nil {
		err = errors.New("lambdaman/ast: no main function")
		return
	}
	offset += len(main.Commands)
	for _, f := range nonmain {
		preamble = append(preamble, Command{
			Name: "LDF",
			Args: []interface{}{
				offset,
			},
			Comment: "load " + f.Name,
		})
		offset += len(f.Commands)
	}
	preamble = append(preamble,
		Command{
			Name: "LDF",
			Args: []interface{}{
				len(preamble) + 3,
			},
			Comment: "load main",
		})
	preamble = append(preamble,
		Command{
			Name: "AP",
			Args: []interface{}{
				len(p.functions) - 1,
			},
		},
		Command{Name: "RTN"},
	)
	preamble.WriteTo(w)
	for _, f := range p.functions {
		fmt.Fprintln(w, ";", f.Name)
		f.Commands.WriteTo(w)
	}
	return
}

type Function struct {
	Name     string
	Commands Commands
	Args     Frame
}

func (f *Function) Add(comment, name string, args ...interface{}) {
	f.Commands = append(f.Commands, Command{
		Name:    name,
		Args:    args,
		Comment: comment,
	})
}

func (f *Function) ResolveSymbol(symbol string) error {
	iframe, idatum, found := f.Args.Find(symbol)
	if !found {
		return errors.New("not found: " + symbol)
	}
	f.Add(symbol, "LD", iframe, idatum)
	return nil
}

type Datum struct {
	Name  string
	Value interface{}
}

type Data []Datum

func (ds Data) Find(name string) (int, bool) {
	for i, d := range ds {
		if d.Name == name {
			return i, true
		}
	}
	return -1, false
}

type Frame struct {
	Data   Data
	Parent *Frame
}

func (f *Frame) Find(name string) (int, int, bool) {
	iframe := 0
	for frame := f; frame != nil; frame = frame.Parent {
		idatum, found := frame.Data.Find(name)
		if found {
			return iframe, idatum, true
		}
		iframe += 1
	}
	return -1, -1, false
}

type Command struct {
	Name    string
	Args    []interface{}
	Comment string
}

func (c Command) String() string {
	raw := c.Name
	if len(c.Args) > 0 {
		raw += " " + fmt.Sprint(c.Args...)
	}
	return raw
}

type Commands []Command

func (cs Commands) WriteTo(w io.Writer) (err error) {
	width := 0
	lines := make([]string, len(cs))
	for i, c := range cs {
		lines[i] = c.String()
		if len(lines[i]) > width {
			width = len(lines[i])
		}
	}
	for i, c := range cs {
		if len(c.Comment) > 0 {
			_, err = fmt.Fprintf(w, "%-*s ; %s\n",
				width, lines[i], c.Comment)
		} else {
			_, err = fmt.Fprintln(w, lines[i])
		}
		if err != nil {
			break
		}
	}
	return
}
