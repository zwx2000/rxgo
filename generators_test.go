package rxgo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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
		})
}

func TestRangeWithCancel(t *testing.T) {
	i := 0

	var oberver = rxgo.ObserverMonitor{}
	oberver.Next = func(y interface{}) {
		x := y.(int)
		if i != x {
			t.Errorf("Ragne Test expect %v but %v", x, i)
		}
		if x >= 5 {
			fmt.Println("End to 5")
			oberver.Unsubscribe()
		}
		i++
	}

	oberver.Context = func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		oberver.CancelObservables = cancel
		//fmt.Println("ctx created ", ctx)
		return ctx
	}

	rxgo.Range(0, 10).Debug(observer{"Range test"}).Subscribe(oberver)
}

func TestGenerator(t *testing.T) {
	i := int64(1)

	//e := errors.New("Flowable Errr")
	// generator function func() (x anytype, end bool)
	rangex := func(start, end int64) func(ctx context.Context) (int64, bool) {
		i := start - 1
		return func(ctx context.Context) (int64, bool) {
			if i < end {
				i++
				if i == 3 {
					//panic(rxgo.FlowableError{Err: e, Elements: nil})
				}
				return i, false
			}
			return 0, true
		}
	}

	rxgo.Start(rangex(1, 5)).Subscribe(
		func(x int64) {
			if i != x {
				t.Errorf("Ragne Test expect %v but %v", x, i)
			}
			i++
		})
}

func TestJust(t *testing.T) {
	i := 10
	rxgo.Just(10, 20, 30).Subscribe(
		func(x int) {
			if i != x {
				t.Errorf("Ragne Test expect %v but %v", x, i)
			}
			i = i + 10
		})
}

func TestFrom(t *testing.T) {
	i := 10
	rxgo.From([]int{10, 20, 30}).Subscribe(
		func(x int) {
			if i != x {
				t.Errorf("From Slice Test expect %v but %v", x, i)
			}
			i = i + 10
		})

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
		})

	i = 10
	ob := rxgo.From([]int{10, 20, 30})
	rxgo.From(ob).Subscribe(
		func(x int) {
			if i != x {
				t.Errorf("From *Observable Test expect %v but %v", x, i)
			}
			i = i + 10
		})
}

func TestThrow(t *testing.T) {
	fail := true
	rxgo.Throw(rxgo.ErrEoFlow).Subscribe(
		rxgo.ObserverMonitor{
			Next: func(x interface{}) {
				t.Errorf("No data expected! but %v", x)
			},
			Error: func(e error) {
				fail = false
				fmt.Println("Find e", e)
				if e != rxgo.ErrEoFlow {
					t.Errorf("Throw Test expect %v but %v", rxgo.ErrEoFlow, e)
				}
			},
			Completed: func() {
				fmt.Println("Completed")
			},
		})
	if fail {
		t.Errorf("No Error !")
	}
}

func TestEmpty(t *testing.T) {
	rxgo.Empty().Subscribe(
		func(x interface{}) {
			t.Errorf("Ragne Test expect %v ", "Nothng")
		})
}

func TestNeverWithCancel(t *testing.T) {

	var oberver = rxgo.ObserverMonitor{
		Next: func(x interface{}) {
			t.Errorf("Ragne Test expect %v ", "Nothng")
		},
	}

	oberver.Context = func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		oberver.CancelObservables = cancel
		fmt.Println("ctx created ", ctx)
		return ctx
	}

	oberver.AfterConnected = func() {
		go func() {
			<-time.After(time.Second * 1)
			fmt.Println("time over!")
			oberver.Unsubscribe()
		}()
	}

	rxgo.Never().Subscribe(oberver)
}
