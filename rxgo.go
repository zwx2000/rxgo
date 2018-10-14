// Copyright 2018 The SS.SYSU Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rxgo provides basic supporting to reactiveX of the Go.
package rxgo

import (
	"errors"
	"reflect"
)

const (
	ThreadingDefault   = 0 // one observable served by one goroutine
	ThreadingIO        = 1 // each item served by one goroutine
	ThreadingComputing = 2 // each item served by one goroutine in a limited group
)

var ErrFuncOnNext = errors.New("Func Error onNext(x)")
var ErrFuncFlip = errors.New("Flip Func Error")

var BufferLen = 128

type Observable struct {
	name      string
	flip      interface{} // transformation function
	outflow   chan interface{}
	operator  func(o *Observable)
	root      *Observable
	next      *Observable
	pred      *Observable
	threading int //threading model
	buf_len   int
	connected bool
	debug     bool
}

func newObservable() *Observable {
	return &Observable{}
}

// connect all Observable form the first one.
func (o *Observable) connect() {
	for po := o.root; po != nil; po = po.next {
		po.Hot()
	}
}

// connect one Observable
func (o *Observable) Hot() *Observable {
	if !o.connected {
		o.operator(o)
	}
	return o
}

func (o *Observable) SubscribeOn(t int) *Observable {
	if !o.connected {
		o.threading = t
	}
	return o
}

func (o *Observable) Subscribe(onNext interface{}, onError func(error), onCompleted func()) {
	fv := reflect.ValueOf(onNext)
	if fv.Kind() != reflect.Func {
		panic(ErrFuncOnNext)
	}

	o.connect()
	//get the last ob servable
	po := o
	for ; po.next != nil; po = po.next {
	}
	for x := range po.outflow {
		params := make([]reflect.Value, 1)
		params[0] = reflect.ValueOf(x)
		fv.Call(params)
	}
}

func (o *Observable) Debug(flag bool) *Observable {
	o.debug = flag
	return o
}
