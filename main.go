package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"lambdaman"
	"os"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "lambdaman.go", nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = lambdaman.Compile(f, os.Stdout)
	if err != nil {
		fmt.Println(err)
	}
}
