// Copyright 2018 The SS.SYSU Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rxgo provides basic supporting to reactiveX of the Go.
package rxgo

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	ThreadingDefault   = 0 // one observable served by one goroutine
	ThreadingIO        = 1 // each item served by one goroutine
	ThreadingComputing = 2 // each item served by one goroutine in a limited group
)

var ErrFuncOnNext = errors.New("Func Error onNext(x)")
var ErrFuncFlip = errors.New("Flip Func Error")

var BufferLen uint = 128

type Observer interface {
	OnNext(x interface{})
	OnError(error)
	OnCompleted()
}

type Observable struct {
	name     string
	flip     interface{} // transformation function
	outflow  chan interface{}
	operator func(o *Observable)
	// chain of Observables
	root *Observable
	next *Observable
	pred *Observable
	// control model
	threading int //threading model. if this is root, it represents obseverOn model
	buf_len   uint
	connected bool
	// utility vars
	debug bool
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
		o.outflow = make(chan interface{}, o.buf_len)
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
	var observer Observer
	if fv.Kind() != reflect.Func {
		// Implements 不能直接使用类型作为参数，导致这种用法非常别扭
		fmt.Printf("onNext v %v t %T \n", onNext, onNext)
		st := reflect.TypeOf((*Observer)(nil)).Elem()
		ft := reflect.TypeOf(onNext)
		fmt.Println("ffffffffffffff", ft, st, ft.Implements(st))
		if ft.Implements(st) {
			observer = onNext.(Observer)
		} else {
			panic(ErrFuncOnNext)
		}
	}

	o.connect()
	//get the last ob servable
	po := o
	for ; po.next != nil; po = po.next {
	}
	for x := range po.outflow {
		if observer != nil {
			observer.OnNext(x)
		} else {
			params := []reflect.Value{reflect.ValueOf(x)}
			fv.Call(params)
		}
	}
}

func (o *Observable) SetBufferLen(length uint) *Observable {
	o.buf_len = length
	return o
}

func (o *Observable) Debug(flag bool) *Observable {
	o.debug = flag
	return o
}
