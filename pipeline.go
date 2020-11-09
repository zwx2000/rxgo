package main

import (
	"fmt"
	RxGo "github.com/zwx2000/rxgo/Rxgo"
)

func main() {
	//define pipeline
	source := RxGo.Just("Hello", "World", "!")
	next := source.Map(func(s string) string {
		return s + " "
	})
	//run pipeline
	next.Subscribe(func(x string) {
		fmt.Print(x)
	})
	fmt.Println()
	source.Subscribe(func(x string) {
		fmt.Print(x)
	})
	fmt.Println()
}
