package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
)

var builtins = map[string]interface{}{
	"append":  nil,
	"cap":     nil,
	"close":   nil,
	"complex": nil,
	"copy":    nil,
	"delete":  nil,
	"imag":    nil,
	"len":     nil,
	"make":    nil,
	"new":     nil,
	"panic":   nil,
	"print":   nil,
	"println": nil,
	"real":    nil,
	"recover": nil,
	"run":     nil,
}

func transpile(inFile string, outFile string) error {

	// Parse the input file.
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, inFile, nil, parser.ParseComments)
	if err != nil {
		log.Fatal("PARSING ERROR: " + err.Error())
	}
	//ast.Print(fset, node)

	// Disallow using raw go.
	ast.Inspect(node, func(n ast.Node) bool {
		if gs, ok := n.(*ast.GoStmt); ok {
			log.Fatal(fmt.Sprintf("go statement not allowed, use run instead (%v)",
				fset.Position(gs.Pos())))
		}
		return true
	})

	var needsContext bool
	var needsRuntime bool

	// All function defintions should use Igor calling convention.
	ast.Inspect(node, func(n ast.Node) bool {
		ft, ok := n.(*ast.FuncType)
		if !ok {
			return true
		}
		ft.Params.List = append([]*ast.Field{ctxArgument}, ft.Params.List...)
		needsContext = true
		return true
	})

	// Add context to all function invocations, except those marked as gocc.
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if id, ok := call.Fun.(*ast.Ident); ok {
			// Transpile "make(nursery)".
			if id.Name == "make" && len(call.Args) == 1 {
				if arg, ok := call.Args[0].(*ast.Ident); ok && arg.Name == "nursery" {
					newNursery(call)
					needsRuntime = true
					return true
				}
			}
			// Built-in functions always use raw calling convention.
			if _, ok = builtins[id.Name]; ok {
				return true
			}
		}
		// Calls explicitly marked with gocc use raw calling convention.
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "gocc" {
				call.Fun = sel.X
				return true
			}
		}
		// Everything else uses Igor calling convention.
		ctx := &ast.Ident{Name: "__ctx"}
		call.Args = append([]ast.Expr{ctx}, call.Args...)
		return true
	})

	// Turn nursery into igor.Nursery.
	ast.Inspect(node, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.Field:
			id, ok := v.Type.(*ast.Ident)
			if ok && id.Name == "nursery" {
				v.Type = typeNursery
				needsRuntime = true
			}
		case *ast.ValueSpec:
			id, ok := v.Type.(*ast.Ident)
			if ok && id.Name == "nursery" {
				v.Type = typeNursery
				needsRuntime = true
			}
		case *ast.TypeSpec:
			id, ok := v.Type.(*ast.Ident)
			if ok && id.Name == "nursery" {
				v.Type = typeNursery
				needsRuntime = true
			}
		case *ast.TypeAssertExpr:
			id, ok := v.Type.(*ast.Ident)
			if ok && id.Name == "nursery" {
				v.Type = typeNursery
				needsRuntime = true
			}
		case *ast.CompositeLit:
			id, ok := v.Type.(*ast.Ident)
			if ok && id.Name == "nursery" {
				v.Type = typeNursery
				needsRuntime = true
			}
		case *ast.ChanType:
			id, ok := v.Value.(*ast.Ident)
			if ok && id.Name == "nursery" {
				v.Value = typeNursery
				needsRuntime = true
			}
		case *ast.ArrayType:
			id, ok := v.Elt.(*ast.Ident)
			if ok && id.Name == "nursery" {
				v.Elt = typeNursery
				needsRuntime = true
			}
		case *ast.MapType:
			id, ok := v.Key.(*ast.Ident)
			if ok && id.Name == "nursery" {
				v.Key = typeNursery
				needsRuntime = true
			}
			id, ok = v.Value.(*ast.Ident)
			if ok && id.Name != "nursery" {
				v.Value = typeNursery
				needsRuntime = true
			}
		}
		return true
	})

	// Expand all run statements.
	ast.Inspect(node, func(n ast.Node) bool {
		blck, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for i, stmt := range blck.List {
			es, ok := stmt.(*ast.ExprStmt)
			if !ok {
				continue
			}
			ce, ok := es.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			id, ok := ce.Fun.(*ast.Ident)
			if !ok {
				continue
			}
			if id.Name != "run" {
				continue
			}
			if len(ce.Args) != 2 {
				log.Fatal(fmt.Sprintf("run should have 2 arguments, has %d (%v)", len(ce.Args),
					fset.Position(ce.Pos())))
			}
			nursery := ce.Args[0]
			goroutine := ce.Args[1]
			ce2, ok := goroutine.(*ast.CallExpr)
			if !ok {
				log.Fatal("unexpected")
			}
			ce2.Args[0] = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "__nursery"},
					Sel: &ast.Ident{Name: "Context__"},
				},
			}
			blck.List[i] = runStatement(nursery, goroutine)
		}
		return true
	})

	// Add main wrapper, as needed.
	hasMain := false
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fn.Name.Name == "main" {
			hasMain = true
			fn.Name.Name = "__main"
		}
		return true
	})
	if hasMain {
		node.Decls = append(node.Decls, mainWrapper)
		needsContext = true
	}

	// Imports needed by the generated code.
	if needsRuntime {
		node.Decls = append([]ast.Decl{importIgor}, node.Decls...)
	}
	if needsContext {
		node.Decls = append([]ast.Decl{importContext}, node.Decls...)
	}

	// Serialize, reformat and save the transpiled code.
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		log.Fatal(err)
	}
	res, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal("FORMATTING ERROR: "+err.Error(), "src", string(buf.Bytes()))
	}
	f, err := os.Create(outFile)
	defer f.Close()
	if _, err = io.WriteString(f, string(res)); err != nil {
		log.Fatal(err.Error())
	}

	return nil
}
