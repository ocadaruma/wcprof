package example

import (
	"fmt"
	"time"
)

// Timer will not be installed to this func
// wcprof: OFF
func privateFunc1() {
	time.Sleep(100 * time.Millisecond)

	fmt.Println("took 100ms")
}

func PublicFunc2() {
	fmt.Println("func2 start")

	for i := 0; i < 10; i++ {
		privateFunc1()
	}

	fmt.Println("took 1000ms")
}

func PublicFunc3() {

	time.Sleep(300 * time.Millisecond)

	fmt.Println("took 300ms")
}
