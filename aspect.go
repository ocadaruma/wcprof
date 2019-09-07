package wcprof

import (
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/goast"
	"github.com/dave/dst/decorator/resolver/guess"
	"github.com/dave/dst/dstutil"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strconv"
	"strings"
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
		dec := decorator.NewDecoratorWithImports(fileset, pkg.Name, goast.New())
		restorer := decorator.NewRestorerWithImports(pkg.Name, guess.New())

		for filename, _ := range pkg.Files {
			dFile, _ := dec.ParseFile(filename, nil, parser.ParseComments)
			applied := dstutil.Apply(dFile, nil, func(cursor *dstutil.Cursor) bool {
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
					}
				}

				return true
			}).(*dst.File)

			err := writeAstToFile(filename, restorer, applied)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// func(t *wcprof.Timer){ t.Stop() }(wcprof.NewTimer("pkg/func name"))
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

func writeAstToFile(filename string, restorer *decorator.Restorer, file *dst.File) error {
	out, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	return restorer.Fprint(out, file)
	//return decorator.Fprint(out, file)
}

//func writeAstToFile(filename string, file *dst.File) error {
//	out, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
//	if err != nil {
//		return err
//	}
//	return decorator.Fprint(out, file)
//}
