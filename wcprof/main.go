package main

import (
	"github.com/ocadaruma/wcprof"
)

func main() {
	//wcprof.InjectTimer(os.Args[1])
	wcprof.InjectTimer("/home/hokada/develop/src/github.com/ocadaruma/wcprof/example", nil)
}
