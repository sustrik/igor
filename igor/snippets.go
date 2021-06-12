package main

import (
	"go/ast"
	"go/token"
)

var importContext = &ast.GenDecl{
	Tok: token.IMPORT,
	Specs: []ast.Spec{
		&ast.ImportSpec{
			Name: &ast.Ident{Name: "__context"},
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"context\"",
			},
		},
	},
}

var importIgor = &ast.GenDecl{
	Tok: token.IMPORT,
	Specs: []ast.Spec{
		&ast.ImportSpec{
			Name: &ast.Ident{Name: "__igor"},
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"github.com/sustrik/igor/lib/igor\"",
			},
		},
	},
}

var typeNursery = &ast.SelectorExpr{
	X:   &ast.Ident{Name: "__igor"},
	Sel: &ast.Ident{Name: "Nursery"},
}

func newNursery(call *ast.CallExpr) {
	call.Fun = &ast.SelectorExpr{
		X:   &ast.Ident{Name: "__igor"},
		Sel: &ast.Ident{Name: "NewNursery"},
	}
	call.Args = []ast.Expr{}
}

var ctxArgument = &ast.Field{
	Names: []*ast.Ident{
		&ast.Ident{Name: "__ctx"},
	},
	Type: &ast.SelectorExpr{
		X:   &ast.Ident{Name: "__context"},
		Sel: &ast.Ident{Name: "Context"},
	},
}

var mainWrapper = &ast.FuncDecl{
	Name: &ast.Ident{Name: "main"},
	Type: &ast.FuncType{},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.Ident{Name: "__main"},
					Args: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "__context"},
								Sel: &ast.Ident{Name: "Background"},
							},
						},
					},
				},
			},
		},
	},
}

func runStatement(nursery ast.Expr, goroutine ast.Expr) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "__nursery"},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					nursery,
				},
			},
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "__nursery"},
						Sel: &ast.Ident{Name: "Start__"},
					},
				},
			},
			&ast.GoStmt{
				Call: &ast.CallExpr{
					Fun: &ast.FuncLit{
						Type: &ast.FuncType{
							Params: &ast.FieldList{},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.AssignStmt{
									Lhs: []ast.Expr{
										&ast.Ident{Name: "__err"},
									},
									Tok: token.DEFINE,
									Rhs: []ast.Expr{
										goroutine,
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "__nursery"},
											Sel: &ast.Ident{Name: "Stop__"},
										},
										Args: []ast.Expr{
											&ast.Ident{Name: "__err"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
