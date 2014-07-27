package ast

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Program struct {
	Name   string
	blocks map[string]*Block
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
	}
	return block
}

func (p *Program) WriteTo(w io.Writer) (err error) {
	var (
		header   = fmt.Sprint("program: ", p.Name)
		preamble = &Block{name: header}
		offset   = len(p.blocks) + 2
		ndecls   = len(p.blocks) - 1
		decls    []Datum
		main     *Block
		nonmain  []*Block
	)

	if ndecls > 0 {
		offset += 1
		preamble.Add("top-level declarations",
			"DUM", ndecls)
	}

	main = p.blocks["main"]
	if main == nil {
		err = errors.New("lambdaman/ast: no main block")
		return
	}

	for _, block := range p.blocks {
		if block.Name() != "main" {
			nonmain = append(nonmain, block)
			decls = append(decls,
				Datum{
					Name: block.Name(),
				})
		}
	}

	for _, block := range nonmain {
		preamble.Add("load "+block.Name(),
			"LDF", block.Name())
	}

	preamble.Add("load main", "LDF", "main")
	rap := "RAP"
	if ndecls == 0 {
		rap = "AP"
	}
	preamble.Add("call main", rap, ndecls)
	preamble.Add("", "RTN")

	main.Env = Frame{Data: decls}

	var all Blocks
	all = preamble.AppendTo(all)
	all = main.AppendTo(all)
	for name, block := range p.blocks {
		if name != "main" {
			block.Env.Parent = &main.Env
			all = block.AppendTo(all)
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

func (b *Block) AppendTo(sequence []*Block) []*Block {
	sequence = append(sequence, b)
	for _, child := range b.Children {
		sequence = child.AppendTo(sequence)
	}
	return sequence
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

func (f *Block) ResolveSymbol(symbol string) error {
	iframe, idatum, found := f.Env.Find(symbol)
	if !found {
		var avail []string
		for frame := &f.Env; frame != nil; frame = frame.Parent {
			for _, datum := range frame.Data {
				avail = append(avail, datum.Name)
			}
		}
		return errors.New(symbol + " not found in: " + strings.Join(avail, " "))
	}
	f.Add(symbol, "LD", iframe, idatum)
	return nil
}

func (b *Block) WriteTo(w io.Writer, lineNums map[string]int) (err error) {
	fmt.Fprintln(w, ";", b.Name())
	cs := make(Commands, len(b.Commands))
	for ic, c := range b.Commands {
		cs[ic], err = c.Compile(&b.Env, lineNums)
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

func (c Command) Compile(f *Frame, lineNums map[string]int) (Command, error) {
	var ok bool
	switch c.Name {
	case "LD":
		if len(c.Args) == 1 {
			name := c.Args[0].(string)
			iframe, idatum, found := f.Find(name)
			if !found {
				return Command{}, errors.New(name)
			}
			c.Args = []interface{}{iframe, idatum}
		}
	case "LDF":
		name := c.Args[0].(string)
		c.Args[0], ok = lineNums[name]
		if !ok {
			return c, errors.New(name + " not found")
		}
	case "SEL":
		name := c.Args[0].(string)
		c.Args[0], ok = lineNums[name]
		if !ok {
			return c, errors.New(name + " not found")
		}
		name = c.Args[1].(string)
		c.Args[1], ok = lineNums[name]
		if !ok {
			return c, errors.New(name + " not found")
		}
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
