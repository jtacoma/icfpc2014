package ast

import (
	"fmt"
	"io"
)

type Program struct {
	Name      string
	Functions map[string]*Function
}

func (p *Program) WriteTo(w io.Writer) (err error) {
	if _, err = fmt.Fprintln(w, "; program:", p.Name); err != nil {
		return
	}
	for _, f := range p.Functions {
		fmt.Fprintln(w, ";", f.Name)
		for _, c := range f.Commands {
			fmt.Fprint(w, c.Name, " ")
			fmt.Fprintln(w, c.Args...)
		}
	}
	return
}

func (p *Program) NewFunction(name string) *Function {
	function := &Function{
		Name: name,
	}
	if p.Functions == nil {
		p.Functions = make(map[string]*Function)
	}
	p.Functions[name] = function
	return function
}

type Function struct {
	Name     string
	Commands []Command
}

func (f *Function) Add(name string, args ...interface{}) {
	f.Commands = append(f.Commands, Command{name, args})
}

type Command struct {
	Name string
	Args []interface{}
}
