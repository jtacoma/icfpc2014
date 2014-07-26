package ast

import (
	"errors"
	"fmt"
	"io"
	"strings"
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
		ndecls   = len(p.functions) - 1
		decls    []Datum
		main     *Function
		nonmain  []*Function
	)
	if ndecls > 0 {
		offset += 1
		preamble = append(preamble,
			Command{
				Name: "DUM",
				Args: []interface{}{
					ndecls,
				},
				Comment: "top-level declarations",
			})
	}
	for i := range p.functions {
		if p.functions[i].Name == "main" {
			main = p.functions[i]
		} else {
			nonmain = append(nonmain, p.functions[i])
			decls = append(decls,
				Datum{
					Name: p.functions[i].Name,
				})
		}
	}
	if main == nil {
		err = errors.New("lambdaman/ast: no main function")
		return
	}
	mainoffset := offset
	offset += len(main.Commands)
	for _, f := range nonmain {
		preamble = append(preamble,
			Command{
				Name: "LDF",
				Args: []interface{}{
					offset,
				},
				Comment: "load " + f.Name,
			})
		offset += len(f.Commands)
	}
	rap := "RAP"
	if ndecls == 0 {
		rap = "AP"
	}
	preamble = append(preamble,
		Command{
			Name: "LDF",
			Args: []interface{}{
				mainoffset,
			},
			Comment: "load main",
		},
		Command{
			Name: rap,
			Args: []interface{}{
				ndecls,
			},
		},
		Command{Name: "RTN"},
	)
	mainframe := Frame{Data: decls}
	preamble.WriteTo(w)
	for ifunc, f := range p.functions {
		fmt.Fprintln(w)
		fmt.Fprintln(w, ";", f.Name)
		if ifunc == 0 {
			f.Args = mainframe
		} else {
			f.Args.Parent = &mainframe
		}
		cs := make(Commands, len(f.Commands))
		for ic, c := range f.Commands {
			cs[ic], err = c.Compile(&f.Args)
			if err != nil {
				return
			}
		}
		cs.WriteTo(w)
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
		var avail []string
		for frame := &f.Args; frame != nil; frame = frame.Parent {
			for _, datum := range frame.Data {
				avail = append(avail, datum.Name)
			}
		}
		return errors.New(symbol + " not found in: " + strings.Join(avail, " "))
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

func (c Command) Compile(f *Frame) (Command, error) {
	if c.Name == "LD" && len(c.Args) == 1 {
		name := c.Args[0].(string)
		iframe, idatum, found := f.Find(name)
		if !found {
			return Command{}, errors.New(name)
		}
		c.Args = []interface{}{iframe, idatum}
	}
	return c, nil
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
