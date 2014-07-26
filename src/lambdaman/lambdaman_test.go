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
package lambdaman
func main() int {
	return 21 + 21
}
`,
		Out: `
; program: lambdaman
LDF 3 ; load main
AP 0
RTN

; main
LDC 21
LDC 21
ADD
RTN
`,
	}, {
		In: `
package lambdaman
func main() int {
	return addtwo(21, 21)
}
func addtwo(a, b int) int {
	return a + b
}
`,
		Out: `
; program: lambdaman
DUM 1  ; top-level declarations
LDF 10 ; load addtwo
LDF 5  ; load main
RAP 1
RTN

; main
LDC 21
LDC 21
LD 0 0 ; addtwo
AP 2
RTN

; addtwo
LD 0 1 ; b
LD 0 0 ; a
ADD
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
