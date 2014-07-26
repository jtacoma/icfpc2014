package main

import (
	"fmt"
	"os"

	"lambdaman/parser"
)

func main() {
	program, err := parser.ParseFile("lambdaman.go", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = program.WriteTo(os.Stdout)
	if err != nil {
		fmt.Println(err)
	}
}
