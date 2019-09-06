package main

import (
	"github.com/ocadaruma/wcprof"
	"os"
)

func main() {
	wcprof.InjectTimer(os.Args[1])
}
