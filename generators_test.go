package rxgo_test

import (
	"testing"

	"github.com/pmlpml/rxgo"
)

func TestRange(t *testing.T) {
	i := 0
	rxgo.Range(0, 10).Subscribe(
		func(x int) {
			if i != x {
				t.Errorf("Ragne Test expect %v but %v", x, i)
			}
			i++
		}, nil, nil)
}

func TestGenerator(t *testing.T) {
	i := int64(1)

	// generator function func() (x anytype, end bool)
	rangex := func(start, end int64) func() (int64, bool) {
		i := start - 1
		return func() (int64, bool) {
			if i < end {
				i++
				return i, false
			}
			return 0, true
		}
	}

	rxgo.Generator(rangex(1, 5)).Subscribe(
		func(x int64) {
			if i != x {
				t.Errorf("Ragne Test expect %v but %v", x, i)
			}
			i++
		}, nil, nil)
}

func TestJust(t *testing.T) {
	i := 10
	rxgo.Just(10, 20, 30).Subscribe(
		func(x int) {
			if i != x {
				t.Errorf("Ragne Test expect %v but %v", x, i)
			}
			i = i + 10
		}, nil, nil)
}

func TestFrom(t *testing.T) {
	i := 10
	rxgo.From([]int{10, 20, 30}).Subscribe(
		func(x int) {
			if i != x {
				t.Errorf("From Slice Test expect %v but %v", x, i)
			}
			i = i + 10
		}, nil, nil)

	i = 10
	ch := make(chan int)
	go func() {
		ch <- 10
		ch <- 20
		ch <- 30
		close(ch)
	}()
	rxgo.From(ch).Subscribe(
		func(x int) {
			if i != x {
				t.Errorf("From Channel Test expect %v but %v", x, i)
			}
			i = i + 10
		}, nil, nil)

	i = 10
	ob := rxgo.From([]int{10, 20, 30})
	rxgo.From(ob).Subscribe(
		func(x int) {
			if i != x {
				t.Errorf("From *Observable Test expect %v but %v", x, i)
			}
			i = i + 10
		}, nil, nil)
}
