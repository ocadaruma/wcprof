package example

import "fmt"

// abc
func PublicTest1() {
	defer func(timer *wcprof.Timer) { timer.Stop() }(wcprof.NewTimer())

	// de
	fmt.Println("hello~~~")
}

func publicTest2() {
	//e
	fmt.Println("world~~~")

	//g
}

// d
func publicTest3() {

}
