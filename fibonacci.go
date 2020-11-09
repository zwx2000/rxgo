package main

import (
	"fmt"
	RxGo "github.com/zwx2000/rxgo/Rxgo"
)

func fibonacci(max int) func() (int, bool) {
	a, b := 0, 1
	return func() (r int, end bool) {
		r = a
		a, b = b, a+b
		if r > max {
			end = true
		}
		return
	}
}

func main() {
	RxGo.Start(fibonacci(10)).Map(func(x int) int {
		return 2*x
	}).Subscribe(func(x int) {
		fmt.Print(x)
	})
}
