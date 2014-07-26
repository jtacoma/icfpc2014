package lambdaman

import (
	gast "go/ast"
	"go/token"
	"io"

	last "lambdaman/ast"
)

func Compile(file *gast.File, w io.Writer) (err error) {
	var program last.Program
	program.Name = file.Name.Name
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *gast.FuncDecl:
			err = CompileFunc(&program, decl)
		}
		if err != nil {
			return
		}
	}
	return program.WriteTo(w)
}

func CompileFunc(program *last.Program, decl *gast.FuncDecl) (err error) {
	function := program.NewFunction(decl.Name.Name)
	for _, stmt := range decl.Body.List {
		switch stmt := stmt.(type) {
		case *gast.ReturnStmt:
			for _, expr := range stmt.Results {
				if err = CompileExpr(function, expr); err != nil {
					return
				}
			}
		}
	}
	return
}

func CompileExpr(function *last.Function, expr gast.Expr) (err error) {
	switch expr := expr.(type) {
	case *gast.BinaryExpr:
		function.Add("LDC", expr.X.(*gast.BasicLit).Value)
		function.Add("LDC", expr.Y.(*gast.BasicLit).Value)
		switch expr.Op {
		case token.ADD:
			function.Add("ADD")
		}
	}
	return
}
