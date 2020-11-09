package main

import (
	"fmt"
	"time"
	rxgo "github.com/zwx2000/rxgo/Rxgo"
)

func main() {
	fmt.Println("test data：1, 2, 3, 4, 5")
	res := []int{}
	ob := rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		return x
	}).Debounce(1000000)
	ob.Subscribe(func(x int) {
		res = append(res, x)
	})
	fmt.Println("After Debounce: ")
	for _, val := range res {
		fmt.Print(val, "  ")
	}
	fmt.Print("\n")

	fmt.Println("test data：1, 2, 3, 4, 5, 2, 3")
		res = []int{}
		ob = rxgo.Just(1, 2, 3, 4, 5, 2, 3).Map(func(x int) int {
			return x
		}).Distinct()
		ob.Subscribe(func(x int) {
			res = append(res, x)
		})
	fmt.Print("After Distinct: ")
	for _, val := range res {
		fmt.Print(val, "  ")
	}
	fmt.Print("\n\n")

	fmt.Println("test data：1, 2, 3, 4, 5")
	res = []int{}
	ob = rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		return x
	}).ElementAt(4)
	ob.Subscribe(func(x int) {
		res = append(res, x)
	})
	fmt.Println("After ElementAt: ", res[0])

	res = []int{}
	ob = rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		return x
	}).First()
	ob.Subscribe(func(x int) {
		res = append(res, x)
	})
	fmt.Println("After First: ", res[0])

	res = []int{}
	ob = rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		return x
	}).Last()
	ob.Subscribe(func(x int) {
		res = append(res, x)
	})
	fmt.Println("After Last: ", res[0])

	res = []int{}
	rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		time.Sleep(2 * time.Millisecond)
		return x
	}).Sample(3 * time.Millisecond).Subscribe(func(x int) {
		res = append(res, x)
	})
	fmt.Print("After Sample: ")
	for _, val := range res {
		fmt.Print(val, "  ")
	}
	fmt.Printf("\n")

	res = []int{}
	ob = rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		return x
	}).Skip(3)
	ob.Subscribe(func(x int) {
		res = append(res, x)
	})
	fmt.Print("After Skip: ")
	for _, val := range res {
		fmt.Print(val, "  ")
	}
	fmt.Print("\n")

	res = []int{}
	ob = rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		return x
	}).SkipLast(3)
	ob.Subscribe(func(x int) {
		res = append(res, x)
	})
	fmt.Print("After SkipLast: ")
	for _, val := range res {
		fmt.Print(val, "  ")
	}
	fmt.Print("\n")

	res = []int{}
	ob = rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		return x
	}).Take(3)
	ob.Subscribe(func(x int) {
		res = append(res, x)
	})
	fmt.Print("After Take: ")
	for _, val := range res {
		fmt.Print(val, "  ")
	}
	fmt.Print("\n")

	res = []int{}
	ob = rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		return x
	}).TakeLast(3)
	ob.Subscribe(func(x int) {
		res = append(res, x)
	})
	fmt.Print("After TakeLast: ")
	for _, val := range res {
		fmt.Print(val, "  ")
	}
	fmt.Print("\n")
}
