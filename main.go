package main

import (
	"golang.org/x/tools/go/ast/astutil"

	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strconv"
)

const (
	MarkerAnnotated  = "wcprof ANNOTATED"
	MarkerOff        = "wcprof OFF"
	ExprProfile      = "func(timer *wcprof.Timer){ timer.Stop() }(wcprof.NewTimer(%s))"
)

func main() {
	fmt.Println("Hello, World")

	fileset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileset, os.Args[1], nil, parser.ParseComments)

	if err != nil {
		log.Fatal(err)
	}
	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				if f, ok := node.(*ast.FuncDecl); ok {
					deferExpr, _ := parser.ParseExpr(
						fmt.Sprintf(ExprProfile, strconv.Quote(pkg.Name + "/" + f.Name.Name)))
					deferStmt := ast.DeferStmt{Call: deferExpr.(*ast.CallExpr)}

					f.Body.List = append([]ast.Stmt{&deferStmt}, f.Body.List...)
				}
				return true
			})

			astutil.AddImport(fileset, file, "github.com/ocadaruma/wcprof")

			err := writeAstToFile(filename, fileset, file)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func writeAstToFile(filename string, fileset *token.FileSet, file *ast.File) error {
	out, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Println(filename)

	return format.Node(out, fileset, file)
}
