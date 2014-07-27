package ast

import (
	"errors"
	"fmt"
	"io"
)

type Program struct {
	Name   string
	blocks map[string]*Block
	consts map[string][]Command
}

func (p *Program) AddConst(name string, value []Command) error {
	if p.consts == nil {
		p.consts = make(map[string][]Command)
	} else if _, dup := p.consts[name]; dup {
		return errors.New("duplicate const name " + name)
	}
	p.consts[name] = value
	return nil
}

func (p *Program) Block(name string) *Block {
	block, exists := p.blocks[name]
	if !exists {
		block = &Block{
			name: name,
		}
		if p.blocks == nil {
			p.blocks = make(map[string]*Block)
		}
		p.blocks[name] = block
		p.AddConst(name, []Command{
			Command{
				Name:    "LDF",
				Args:    []interface{}{name},
				Comment: name,
			},
		})
	}
	return block
}

func (p *Program) WriteTo(w io.Writer) (err error) {
	var (
		main *Block
		all  Blocks
	)

	main = p.blocks["main"]
	if main == nil {
		err = errors.New("lambdaman/ast: no main block")
		return
	}
	all, err = main.ExpandTo(all, p.consts)
	if err != nil {
		return
	}

	for _, block := range p.blocks {
		if block.Name() != "main" {
			all, err = block.ExpandTo(all, p.consts)
			if err != nil {
				return
			}
		}
	}

	lineNums := all.LineNums()
	for iblock, block := range all {
		if iblock != 0 {
			fmt.Fprintln(w)
		}
		err = block.WriteTo(w, lineNums)
		if err != nil {
			break
		}
	}
	return
}

type Blocks []*Block

func (blocks Blocks) LineNums() map[string]int {
	var (
		line     = 0
		lineNums = make(map[string]int)
	)
	for _, block := range blocks {
		lineNums[block.Name()] = line
		line += len(block.Commands)
	}
	return lineNums
}

type Block struct {
	name     string
	Commands Commands
	Env      Frame
	Children []*Block
}

func (b *Block) Name() string { return b.name }

func (b *Block) ExpandTo(sequence []*Block, consts map[string][]Command) ([]*Block, error) {
	denamed := &Block{
		name: b.name,
	}
	for _, cmd := range b.Commands {
		evaled, err := cmd.EvalNames(&b.Env, consts)
		if err != nil {
			if nf, ok := err.(*NotFoundError); ok {
				nf.BlockName = b.name
			}
			return nil, err
		}
		denamed.Commands = append(denamed.Commands, evaled...)
	}
	sequence = append(sequence, denamed)
	for _, child := range b.Children {
		var err error
		sequence, err = child.ExpandTo(sequence, consts)
		if err != nil {
			if nf, ok := err.(*NotFoundError); ok {
				nf.BlockName = b.name
			}
			return nil, err
		}
	}
	return sequence, nil
}

func (b *Block) Child(affix string) *Block {
	prefix := fmt.Sprintf("%s.%d",
		b.Name(), len(b.Commands))
	child := &Block{
		name: prefix + affix,
		Env:  b.Env,
	}
	b.Children = append(b.Children, child)
	return child
}

func (b *Block) Add(comment, name string, args ...interface{}) {
	b.Commands = append(b.Commands, Command{
		Name:    name,
		Args:    args,
		Comment: comment,
	})
}

func (b *Block) WriteTo(w io.Writer, lineNums map[string]int) (err error) {
	fmt.Fprintln(w, ";", b.Name())
	cs := make(Commands, len(b.Commands))
	for ic, c := range b.Commands {
		cs[ic], err = c.EvalLineNums(lineNums)
		if err != nil {
			return
		}
	}
	cs.WriteTo(w)
	return
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

func (c Command) EvalNames(f *Frame, consts map[string][]Command) ([]Command, error) {
	var (
		ok   bool
		cmds []Command
	)
	switch c.Name {
	default:
		cmds = append(cmds, c)
	case "LD":
		if len(c.Args) == 1 {
			name := c.Args[0].(string)
			iframe, idatum, found := f.Find(name)
			if !found {
				cmds, found = consts[name]
				if !found {
					return nil,
						&NotFoundError{VarName: name}
				}
			} else {
				c.Args = []interface{}{iframe, idatum}
				cmds = append(cmds, c)
			}
		}
	case "LDF":
		name := c.Args[0].(string)
		cmds, ok = consts[name]
		if !ok {
			// pass through, maybe EvalLineNums will resolve it
			cmds = []Command{c}
		}
	}
	return cmds, nil
}

func (c Command) EvalLineNums(lineNums map[string]int) (Command, error) {
	var ok bool
	switch c.Name {
	case "LDF":
		name := c.Args[0].(string)
		c.Args[0], ok = lineNums[name]
		if !ok {
			return c, &NotFoundError{VarName: name}
		}
	case "SEL":
		name := c.Args[0].(string)
		c.Args[0], ok = lineNums[name]
		if !ok {
			return c, &NotFoundError{VarName: name}
		}
		name = c.Args[1].(string)
		c.Args[1], ok = lineNums[name]
		if !ok {
			return c, &NotFoundError{VarName: name}
		}
	}
	return c, nil
}

func (c Command) SetComment(comment string) Command {
	return Command{
		Name:    c.Name,
		Args:    c.Args,
		Comment: comment,
	}
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
