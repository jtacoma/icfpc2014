package lambdaman

import (
	"ghc"
)

func main(addtwo func(int, int) int) int {
	return addtwo(10, 32)
}

func addtwo(a, b int) int {
	return a + b
}
