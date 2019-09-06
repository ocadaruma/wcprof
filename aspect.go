package wcprof

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	Marker       = "// wcprof: MARKED"
	MarkerOff    = "// wcprof: OFF"
	ExprProfile  = "func(timer *wcprof.Timer){ timer.Stop() }(wcprof.NewTimer(%s))"
)

func InjectTimer(filepath string) {
	fileset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileset, filepath, nil, parser.ParseComments)

	if err != nil {
		log.Fatal(err)
	}
	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				if f, ok := node.(*ast.FuncDecl); ok {
					marked := false
					if f.Doc != nil {
						for _, doc := range f.Doc.List {
							if strings.Contains(doc.Text, Marker) || strings.Contains(doc.Text, MarkerOff) {
								marked = true
								break
							}
						}
					}

					if !marked {
						if f.Doc == nil {
							f.Doc = &ast.CommentGroup{}
						}
						f.Doc.List = append(f.Doc.List, &ast.Comment{Text: Marker})
						deferExpr, _ := parser.ParseExpr(
							fmt.Sprintf(ExprProfile, strconv.Quote(pkg.Name + "/" + f.Name.Name)))
						deferStmt := ast.DeferStmt{Call: deferExpr.(*ast.CallExpr)}

						f.Body.List = append([]ast.Stmt{&deferStmt}, f.Body.List...)
					}
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
	return format.Node(out, fileset, file)
}
