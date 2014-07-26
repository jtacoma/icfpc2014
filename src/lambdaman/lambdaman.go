package lambdaman

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
)

func Compile(file *ast.File, w io.Writer) (err error) {
	fmt.Fprintln(w, "; package", file.Name)
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			fmt.Fprintln(w, ";", decl.Name)
			if err = CompileBlock(decl.Body, w); err != nil {
				return
			}
		}
	}
	return
}

func CompileBlock(body *ast.BlockStmt, w io.Writer) (err error) {
	for _, stmt := range body.List {
		switch stmt := stmt.(type) {
		case *ast.ReturnStmt:
			for _, expr := range stmt.Results {
				if err = CompileExpr(expr, w); err != nil {
					return
				}
			}
		}
	}
	return
}

func CompileExpr(expr ast.Expr, w io.Writer) (err error) {
	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		fmt.Fprintln(w, "LDC", expr.X.(*ast.BasicLit).Value)
		fmt.Fprintln(w, "LDC", expr.Y.(*ast.BasicLit).Value)
		switch expr.Op {
		case token.ADD:
			fmt.Fprintln(w, "ADD")
		}
	}
	return
}
