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
; program: branching
LDF 3 ; load main
AP 0  ; call main
RTN

; main
LDC 1
LDC 0
CGTE
SEL 8 12 ; main.3t main.3f
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
; program: calling
DUM 2  ; top-level declarations
LDF 12 ; load arithmetic
LDF 19 ; load addtwo
LDF 6  ; load main
RAP 2  ; call main
RTN

; main
LDC 1
LDC 2
LDC 3
LD 0 0 ; arithmetic
AP 3
RTN

; arithmetic
LD 0 0 ; a
LD 0 1 ; b
LD 1 1 ; addtwo
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
func main() {
	[]interface{}{
		42,
		step,
	}
}
func step(s int) {
	[]interface{}{
		s + 1,
		down,
	}
}
`,
		Out: `
; program: lambdaman
DUM 1 ; top-level declarations
LDF 9 ; load step
LDF 5 ; load main
RAP 1 ; call main
RTN

; main
LDC 42
LD 0 0 ; step
CONS
RTN

; step
LD 0 0 ; s
LDC 1
ADD
LDC 2  ; down
CONS
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
