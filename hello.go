package main

import (
	"fmt"
	RxGo "github.com/zwx2000/rxgo/Rxgo"
)

func main() {
	RxGo.Just("Hello", "World", "!").Subscribe(func(x string) {
		fmt.Println(x)
	})
}
