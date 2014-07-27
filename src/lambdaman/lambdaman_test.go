package lambdaman

import (
	"bytes"
	"strings"
	"testing"

	"lambdaman/parser"
)

var tests = []struct {
	In  string
	Out string
}{
	{
		In: `
package branching
func main() int {
	if 0 <= 1 {
		21 * 2
	} else {
		42 / 2
	}
}
`,
		Out: `
; main
LDC 1
LDC 0
CGTE
SEL 5 9 ; main.3t main.3f
RTN

; main.3t
LDC 21
LDC 2
MUL
JOIN

; main.3f
LDC 42
LDC 2
DIV
JOIN
`,
	}, {
		In: `
package calling
func main() int {
	arithmetic(1, 2, 3)
}
func arithmetic(a, b, c int) int {
	addtwo(a, b) * c
}
func addtwo(a, b int) int {
	a + b
}
`,
		Out: `
; main
LDC 1
LDC 2
LDC 3
LDF 6 ; arithmetic
AP 3
RTN

; arithmetic
LD 0 0 ; a
LD 0 1 ; b
LDF 13 ; addtwo
AP 2
LD 0 2 ; c
MUL
RTN

; addtwo
LD 0 0 ; a
LD 0 1 ; b
ADD
RTN
`,
	}, {
		In: `
package lambdaman
const (
	up = 0
	right = 1
	down = 2
	left = 3
)
func main(world, ghosts interface{}) {
	[]interface{}{
		NewMem(world, ghosts),
		step,
	}
}
func step(mem, world interface{}) {
	split(NextMem(mem, world))
}
func split(mem interface{}) {
	[]interface{}{
		mem,
		Direction(mem),
	}
}
func NewMem(world, ghosts interface{}) {
	42
}
func NextMem(mem, world interface{}) {
	mem + 1
}
func Direction(mem interface{}) {
	mem - (mem / 4) * 4
}
`,
		Out: `
; main
LD 0 0 ; world
LD 0 1 ; ghosts
LDF 20 ; NewMem
AP 2
LDF 7  ; step
CONS
RTN

; step
LD 0 0 ; mem
LD 0 1 ; world
LDF 22 ; NextMem
AP 2
LDF 14 ; split
AP 1
RTN

; split
LD 0 0 ; mem
LD 0 0 ; mem
LDF 26 ; Direction
AP 1
CONS
RTN

; NewMem
LDC 42
RTN

; NextMem
LD 0 0 ; mem
LDC 1
ADD
RTN

; Direction
LD 0 0 ; mem
LD 0 0 ; mem
LDC 4
DIV
LDC 4
MUL
SUB
RTN
`,
	},
}

func diff(a, b string) []string {
	var result []string
	as := strings.Split(a, "\n")
	bs := strings.Split(b, "\n")
	numlines := len(as)
	if numlines < len(bs) {
		numlines = len(bs)
	}
	for i := 0; i < numlines; i += 1 {
		var aline, bline string
		if i < len(as) {
			aline = as[i]
		}
		if i < len(bs) {
			bline = bs[i]
		}
		if aline != bline {
			if len(aline) > 0 {
				result = append(result, "- "+aline)
			}
			if len(bline) > 0 {
				result = append(result, "+ "+bline)
			}
		}
	}
	return result
}

func TestCompile(t *testing.T) {
	for itest, test := range tests {
		program, err := parser.ParseFile("src.go", test.In)
		if err != nil {
			t.Errorf("test#%d: %s", itest, err)
			continue
		}
		var buffer bytes.Buffer
		err = program.WriteTo(&buffer)
		if err != nil {
			t.Errorf("test#%d: %s", itest, err)
			continue
		}
		a := strings.TrimSpace(test.Out)
		b := strings.TrimSpace(buffer.String())
		for _, delta := range diff(a, b) {
			t.Errorf("test#%d: %q", itest, delta)
		}
	}
}
