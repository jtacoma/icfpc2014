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

func TransformGoFunc(p *ast.Program, decl *gast.FuncDecl) error {
	block := p.Block(decl.Name.Name)
	err := appendGoStmts(block, decl.Body.List)
	if err != nil {
		return err
	}
	block.Add("", "RTN")
	for _, param := range decl.Type.Params.List {
		for _, ident := range param.Names {
			block.Env.Data = append(block.Env.Data,
				ast.Datum{Name: ident.Name})
		}
	}
	return nil
}

func appendGoStmts(b *ast.Block, stmts []gast.Stmt) (err error) {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *gast.ExprStmt:
			err = TransformGoExpr(b, stmt.X)
			if err != nil {
				return
			}
		case *gast.IfStmt:
			err = TransformGoExpr(b, stmt.Cond)
			tbranch := b.Child("t")
			appendGoStmts(tbranch, stmt.Body.List)
			tbranch.Add("", "JOIN")
			fbranch := b.Child("f")
			if stmt.Else != nil {
				appendGoStmts(fbranch,
					stmt.Else.(*gast.BlockStmt).List)
			}
			fbranch.Add("", "JOIN")
			comment := tbranch.Name() + " " + fbranch.Name()
			b.Add(comment, "SEL",
				tbranch.Name(), fbranch.Name())
		case *gast.ReturnStmt:
			return errors.New("explicit return not supported")
		}
	}
	return
}

func TransformGoExpr(block *ast.Block, expr gast.Expr) (err error) {
	switch expr := expr.(type) {
	case *gast.BasicLit:
		block.Add("", "LDC", expr.Value)
	case *gast.BinaryExpr:
		switch expr.Op {
		case gtoken.ADD:
			TransformGoExpr(block, expr.X)
			TransformGoExpr(block, expr.Y)
			block.Add("", "ADD")
		case gtoken.SUB:
			TransformGoExpr(block, expr.X)
			TransformGoExpr(block, expr.Y)
			block.Add("", "SUB")
		case gtoken.MUL:
			TransformGoExpr(block, expr.X)
			TransformGoExpr(block, expr.Y)
			block.Add("", "MUL")
		case gtoken.QUO:
			TransformGoExpr(block, expr.X)
			TransformGoExpr(block, expr.Y)
			block.Add("", "DIV")
		case gtoken.EQL: // ==
			TransformGoExpr(block, expr.X)
			TransformGoExpr(block, expr.Y)
			block.Add("", "CEQ")
		case gtoken.LSS: // <
			TransformGoExpr(block, expr.Y)
			TransformGoExpr(block, expr.X)
			block.Add("", "CGT")
		case gtoken.GTR: // >
			TransformGoExpr(block, expr.X)
			TransformGoExpr(block, expr.Y)
			block.Add("", "CGT")
		case gtoken.NEQ: // !=
			TransformGoExpr(block, expr.X)
			TransformGoExpr(block, expr.Y)
			block.Add("", "CEQ")
			block.Add("", "LDC", 0)
			block.Add("", "CEQ")
		case gtoken.LEQ: // <=
			TransformGoExpr(block, expr.Y)
			TransformGoExpr(block, expr.X)
			block.Add("", "CGTE")
		case gtoken.GEQ: // >=
			TransformGoExpr(block, expr.X)
			TransformGoExpr(block, expr.Y)
			block.Add("", "CGTE")
		default:
			err = errors.New("unsupported operation")
		}
	case *gast.Ident:
		block.Add(expr.Name, "LD", expr.Name)
	case *gast.CallExpr:
		for _, arg := range expr.Args {
			err = TransformGoExpr(block, arg)
			if err != nil {
				return
			}
		}
		err = TransformGoExpr(block, expr.Fun)
		block.Add("", "AP", len(expr.Args))
	}
	return
}
