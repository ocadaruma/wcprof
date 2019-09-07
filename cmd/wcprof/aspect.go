package main

import (
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/goast"
	"github.com/dave/dst/decorator/resolver/guess"
	"github.com/dave/dst/dstutil"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	WcprofPath  = "github.com/ocadaruma/wcprof"
	Marker      = "// wcprof: MARKED"
	MarkerOff   = "// wcprof: OFF"
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

func InterceptTimer(filepath string, config *Config) {
	if config == nil {
		config = DefaultConfig
	}

	fileset := token.NewFileSet()

	// actual file parsing is done by dst below
	pkgs, err := parser.ParseDir(fileset, filepath, config.Filter, parser.PackageClauseOnly)

	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range pkgs {
		decor := decorator.NewDecoratorWithImports(fileset, pkg.Name, goast.New())
		restorer := decorator.NewRestorerWithImports(pkg.Name, guess.New())

		for filename, _ := range pkg.Files {
			file, err := decor.ParseFile(filename, nil, parser.ParseComments)
			if err != nil {
				log.Fatal(err)
			}

			modified := false
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
						f.Body.List = append([]dst.Stmt{timerDefer(pkg.Name, f.Name.Name)}, f.Body.List...)
						modified = true
					}
				}

				return true
			}).(*dst.File)

			if modified {
				err = writeAstToFile(filename, restorer, applied, config.Backup)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

// Generate AST that measuring method execution time
// Actual code look like this:
//   func(t *wcprof.Timer){ t.Stop() }(wcprof.NewTimer("pkg/func name"))
func timerDefer(pkg, funcName string) *dst.DeferStmt {
	return &dst.DeferStmt{
		Call: &dst.CallExpr{
			Fun: &dst.FuncLit{
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							&dst.Field{
								Names: []*dst.Ident{
									&dst.Ident{Name: "t"},
								},
								Type: &dst.StarExpr{
									X: &dst.Ident{
										Name: "Timer",
										Path: WcprofPath,
									},
								},
							},
						},
					},
				},
				Body: &dst.BlockStmt{
					List: []dst.Stmt{
						&dst.ExprStmt{
							X: &dst.CallExpr{
								Fun: &dst.SelectorExpr{
									X: &dst.Ident{
										Name: "t",
									},
									Sel: &dst.Ident{
										Name: "Stop",
									},
								},
							},
						},
					},
				},
			},
			Args: []dst.Expr{
				&dst.CallExpr{
					Fun: &dst.Ident{
						Name: "NewTimer",
						Path: WcprofPath,
					},
					Args: []dst.Expr{
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: strconv.Quote(pkg + "/" + funcName),
						},
					},
				},
			},
		},
	}
}

func writeAstToFile(filename string, restorer *decorator.Restorer, file *dst.File, backup bool) error {
	if backup {
		backupfile := filename + "_" + time.Now().Format("20060102150405")
		src, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer src.Close()

		dest, err := os.Create(backupfile)
		if err != nil {
			log.Fatal(err)
		}
		defer dest.Close()

		_, err = io.Copy(dest, src)
		if err != nil {
			log.Fatal(err)
		}
	}

	out, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	return restorer.Fprint(out, file)
}
