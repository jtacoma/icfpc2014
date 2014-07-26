package parser

import (
	gast "go/ast"
	gparser "go/parser"
	gtoken "go/token"

	"lambdaman/ast"
)

func ParseFile(filename string, file interface{}) (*ast.Program, error) {
	var (
		err     error
		go_fset *gtoken.FileSet
		go_file *gast.File
	)
	go_fset = gtoken.NewFileSet()
	go_file, err = gparser.ParseFile(go_fset, filename, file, 0)
	if err != nil {
		return nil, err
	}
	return TransformGoFile(go_file)
}

func TransformGoFile(go_file *gast.File) (program *ast.Program, err error) {
	program = &ast.Program{
		Name: go_file.Name.Name,
	}
	for _, decl := range go_file.Decls {
		switch decl := decl.(type) {
		case *gast.FuncDecl:
			err = TransformGoFunc(program, decl)
		}
		if err != nil {
			return nil, err
		}
	}
	return program, nil
}

func TransformGoFunc(p *ast.Program, decl *gast.FuncDecl) (err error) {
	function := p.NewFunction(decl.Name.Name)
	for _, stmt := range decl.Body.List {
		switch stmt := stmt.(type) {
		case *gast.ReturnStmt:
			TransformGoReturn(function, stmt)
		}
	}
	return
}

func TransformGoReturn(f *ast.Function, stmt *gast.ReturnStmt) (err error) {
	for _, expr := range stmt.Results {
		err = TransformGoExpr(f, expr)
		if err != nil {
			return
		}
	}
	return
}

func TransformGoExpr(function *ast.Function, expr gast.Expr) (err error) {
	switch expr := expr.(type) {
	case *gast.BinaryExpr:
		function.Add("LDC", expr.X.(*gast.BasicLit).Value)
		function.Add("LDC", expr.Y.(*gast.BasicLit).Value)
		switch expr.Op {
		case gtoken.ADD:
			function.Add("ADD")
		}
	}
	return
}
