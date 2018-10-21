// Copyright 2018 The SS.SYSU Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rxgo provides basic supporting to reactiveX of the Go.
package rxgo

import (
	"context"
	"errors"
	"reflect"
)

type ThreadModel uint

const (
	ThreadingDefault   ThreadModel = iota // one observable served by one goroutine
	ThreadingIO                           // each item served by one goroutine
	ThreadingComputing                    // each item served by one goroutine in a limited group
)

// Subscribe paeameter error
var ErrFuncOnNext = errors.New("Subscribe paramteter needs func(x,anytype) or Observer or ObserverWithContext")

// operator func error
var ErrFuncFlip = errors.New("Operator Func Error")

// if user function throw EoFlow, the Observeable will stop and close it
var ErrEoFlow = errors.New("End of Flow!")

// if user function throw SkipItem, the Observeable will skip current item
var ErrSkipItem = errors.New("Skip item!")

// Error that can flow to subscriber or user function which processes error as an input
type FlowableError struct {
	Err      error
	Elements interface{}
}

func (e FlowableError) Error() string {
	return e.Err.Error()
}

// default buffer of channels
var BufferLen uint = 128

// Observer subscribes to an Observable. Then that observer reacts to whatever item or sequence of items the Observable emits.
type Observer interface {
	OnNext(x interface{})
	OnError(error)
	OnCompleted()
}

// Make Observables Context and support unsubscribe operation
type ObserverWithContext interface {
	Observer
	GetObserverContext() context.Context // you must create a cancelable context here when unsubscribe
	OnConnected()
	Unsubscribe()
}

// Create observer quickly with function
type ObserverMonitor struct {
	Next              func(x interface{})
	Error             func(error)
	Completed         func()
	Context           func() context.Context // an observer context musit gived when observables before connected
	AfterConnected    func()
	CancelObservables context.CancelFunc
}

func (o ObserverMonitor) OnNext(x interface{}) {
	if o.Next != nil {
		o.Next(x)
	}
}

func (o ObserverMonitor) OnError(e error) {
	if o.Error != nil {
		o.Error(e)
	}
}

func (o ObserverMonitor) OnCompleted() {
	if o.Completed != nil {
		o.Completed()
	}
}

func (o ObserverMonitor) GetObserverContext() (c context.Context) {
	if o.Context != nil {
		return o.Context()
	}
	return context.Background()
}

func (o ObserverMonitor) OnConnected() {
	if o.AfterConnected != nil {
		o.AfterConnected()
	}
}

func (o ObserverMonitor) Unsubscribe() {
	if o.CancelObservables != nil {
		o.CancelObservables()
	}
}

// An Observable is a 'collection of items that arrive over time'. Observables can be used to model asynchronous events.
// Observables can also be chained by operators to transformed, combined those items
// The Observable's operators, by default, run with a channel size of 128 elements except that the source (first) observable has no buffer
type Observable struct {
	name     string
	flip     interface{} // transformation function
	outflow  chan interface{}
	operator func(ctx context.Context, o *Observable)
	// chain of Observables
	root *Observable
	next *Observable
	pred *Observable
	// control model
	threading ThreadModel //threading model. if this is root, it represents obseverOn model
	buf_len   uint
	connected bool
	// utility vars
	debug             Observer
	flip_sup_ctx      bool //indicate that flip function use context as first paramter
	flip_accept_error bool // indicate that flip function input's data is type interface{} or error
}

func newObservable() *Observable {
	return &Observable{}
}

// connect all Observable form the first one.
func (o *Observable) connect(ctx context.Context) {
	for po := o.root; po != nil; po = po.next {
		po.Hot(ctx)
	}
}

// connect one Observable
func (o *Observable) Hot(ctx context.Context) *Observable {
	if !o.connected {
		o.outflow = make(chan interface{}, o.buf_len)
		o.connected = true
		o.operator(ctx, o)
	}
	return o
}

func (o *Observable) SubscribeOn(t ThreadModel) *Observable {
	if !o.connected {
		o.threading = t
	}
	return o
}

func (o *Observable) ObserveOn(t ThreadModel) *Observable {
	if !o.connected {
		po := o.root
		po.threading = t
	}
	return o
}

func (o *Observable) Subscribe(ob interface{}) {
	fv, ft := reflect.ValueOf(ob), reflect.TypeOf(ob)

	var observer Observer

	// observe function `func(x anytype)`
	if fv.Kind() == reflect.Func {
		if ft.NumIn() == 1 && ft.NumOut() != 0 {
			panic(ErrFuncOnNext)
		}
	} else {
		st := reflect.TypeOf((*Observer)(nil)).Elem() // get type of *Observer
		//fmt.Println("ffffffffffffff", ft, st, ft.Implements(st))
		if ft.Implements(st) {
			observer = ob.(Observer)
		} else {
			panic(ErrFuncOnNext)
		}
	}

	oc, ctxok := observer.(ObserverWithContext)
	ctx := context.Background()

	if ctxok {
		ctx = oc.GetObserverContext()
		//fmt.Println("ctx geted!", ctx)
	}

	o.connect(ctx)
	if ctxok {
		oc.OnConnected()
	}

	//get the last ob servable
	po := o
	for ; po.next != nil; po = po.next {
	}
	for x := range po.outflow {
		if observer != nil {
			if e, ok := x.(error); ok {
				observer.OnError(e)

			} else {
				observer.OnNext(x)
			}
		} else {
			if _, ok := x.(error); ok {
				// skip error
			} else {
				params := []reflect.Value{reflect.ValueOf(x)}
				fv.Call(params)
			}
		}
	}
	if observer != nil {
		observer.OnCompleted()
	}
}

func (o *Observable) SetBufferLen(length uint) *Observable {
	o.buf_len = length
	return o
}

func (o *Observable) Debug(monitor Observer) *Observable {
	o.debug = monitor
	return o
}

func (o *Observable) closeFlow() *Observable {
	if !o.connected {
		//Todo
		return o
	}
	// maybe need waiting for parent observable closed
	close(o.outflow)
	if o.debug != nil {
		o.debug.OnCompleted()
	}
	o.connected = false
	return o
}
