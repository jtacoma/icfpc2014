package parser

import (
	"errors"
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
	topdecls := ast.Frame{}
	for _, decl := range go_file.Decls {
		switch decl := decl.(type) {
		case *gast.FuncDecl:
			if decl.Name.Name != "main" {
				topdecls.Data = append(topdecls.Data,
					ast.Datum{Name: decl.Name.Name})
			}
		}
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
	for _, param := range decl.Type.Params.List {
		for _, ident := range param.Names {
			function.Args.Data = append(function.Args.Data,
				ast.Datum{Name: ident.Name})
		}
	}
	for _, stmt := range decl.Body.List {
		switch stmt := stmt.(type) {
		case *gast.ExprStmt:
			err = TransformGoExpr(function, stmt.X)
			if err != nil {
				return
			}
		case *gast.ReturnStmt:
			for _, expr := range stmt.Results {
				err = TransformGoExpr(function, expr)
				if err != nil {
					return
				}
			}
			function.Add("", "RTN")
		}
	}
	return
}

func TransformGoExpr(function *ast.Function, expr gast.Expr) (err error) {
	switch expr := expr.(type) {
	case *gast.BasicLit:
		function.Add("", "LDC", expr.Value)
	case *gast.BinaryExpr:
		switch expr.Op {

		case gtoken.ADD:
			TransformGoExpr(function, expr.X)
			TransformGoExpr(function, expr.Y)
			function.Add("", "ADD")
		case gtoken.SUB:
			TransformGoExpr(function, expr.X)
			TransformGoExpr(function, expr.Y)
			function.Add("", "SUB")
		case gtoken.MUL:
			TransformGoExpr(function, expr.X)
			TransformGoExpr(function, expr.Y)
			function.Add("", "MUL")
		case gtoken.QUO:
			TransformGoExpr(function, expr.X)
			TransformGoExpr(function, expr.Y)
			function.Add("", "DIV")

		case gtoken.EQL: // ==
			TransformGoExpr(function, expr.X)
			TransformGoExpr(function, expr.Y)
			function.Add("", "CEQ")
		case gtoken.LSS: // <
			TransformGoExpr(function, expr.Y)
			TransformGoExpr(function, expr.X)
			function.Add("", "CGT")
		case gtoken.GTR: // >
			TransformGoExpr(function, expr.X)
			TransformGoExpr(function, expr.Y)
			function.Add("", "CGT")

		case gtoken.NEQ: // !=
			TransformGoExpr(function, expr.X)
			TransformGoExpr(function, expr.Y)
			function.Add("", "CEQ")
			function.Add("", "LDC", 0)
			function.Add("", "CEQ")
		case gtoken.LEQ: // <=
			TransformGoExpr(function, expr.Y)
			TransformGoExpr(function, expr.X)
			function.Add("", "GEQ")
		case gtoken.GEQ: // >=
			TransformGoExpr(function, expr.X)
			TransformGoExpr(function, expr.Y)
			function.Add("", "GEQ")

		default:
			err = errors.New("unsupported operation")
		}
	case *gast.Ident:
		function.Add(expr.Name, "LD", expr.Name)
	case *gast.CallExpr:
		for _, arg := range expr.Args {
			err = TransformGoExpr(function, arg)
			if err != nil {
				return
			}
		}
		err = TransformGoExpr(function, expr.Fun)
		function.Add("", "AP", len(expr.Args))
	}
	return
}
