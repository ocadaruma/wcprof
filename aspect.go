package wcprof

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	WcprofPkg    = "github.com/ocadaruma/wcprof"
	Marker       = "// wcprof: MARKED"
	MarkerOff    = "// wcprof: OFF"
	ExprProfile  = "func(t *wcprof.Timer){ t.Stop() }(wcprof.NewTimer(%s))"
)

type Config struct {
	Filter func(file os.FileInfo) bool
	Backup bool
}

var DefaultConfig = &Config{
	Filter: func(file os.FileInfo) bool {
		return !strings.HasSuffix(file.Name(), "test.go")
	},
	Backup: false,
}

func InjectTimer(filepath string, config *Config) {
	if config == nil {
		config = DefaultConfig
	}

	fileset := token.NewFileSet()
	pkgs, err := decorator.ParseDir(fileset, filepath, config.Filter, parser.ParseComments)

	if err != nil {
		log.Fatal(err)
	}
	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			applied := dstutil.Apply(file, nil, func(cursor *dstutil.Cursor) bool {
				if f, ok := cursor.Node().(*dst.FuncDecl); ok {
					marked := false
					if f.Decorations() != nil {
						for _, dec := range f.Decorations().Start {
							if strings.Contains(dec, Marker) || strings.Contains(dec, MarkerOff) {
								marked = true
								break
							}
						}
					}

					if !marked {
						f.Decs.Start.Append(Marker)

						deferExpr, _ := parser.ParseExpr(
							fmt.Sprintf(ExprProfile, strconv.Quote(pkg.Name + "/" + f.Name.Name)))
						deferStmt := ast.DeferStmt{Call: deferExpr.(*ast.CallExpr)}
						decoratedDefer, _ := decorator.Decorate(fileset, &deferStmt)

						f.Body.List = append([]dst.Stmt{decoratedDefer.(*dst.DeferStmt)}, f.Body.List...)
					}
				}

				return true
			}).(*dst.File)

			err := writeAstToFile(filename, applied)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func writeAstToFile(filename string, file *dst.File) error {
	out, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	return decorator.Fprint(out, file)
}
