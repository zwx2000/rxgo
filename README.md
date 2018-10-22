# RxGo: Reactive Extensions for the Go
a go reactiveX implementation  ref  

RxGo is a Go implementation of [Reactive Extensions](http://reactivex.io/documentation/observable.html): a library for composing asynchronous and event-based programs by using observable sequences.

It extends the observer pattern to support sequences of data/events and adds operators that allow you to compose sequences together declaratively while abstracting away concerns about things like low-level threading, synchronization, thread-safety and concurrent data structures.

#### V0.2

* function program friendly
* Connectable observables, support Hot or Cold item emiting
* context support for concurrency
* support error type data in observable sequences
* unsubscribe supporting
* non-opinionated about source of concurrency 
* async or synchronous execution

## Getting started

### Hello World
Tthe **Hello World** program:

```go
package main

import (
	"fmt"
	RxGo "github.com/pmlpml/rxgo"
)

func main() {
	RxGo.Just("Hello", "World", "!").Subscribe(func(x string) {
		fmt.Println(x)
	})
}
```

output:

```
Hello
World
!
```

