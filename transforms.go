// Copyright 2018 The SS.SYSU Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rxgo

import (
	"fmt"
	"reflect"
	"sync"
)

// Map maps a MappableFunc `func(x anytype) anytype` predicate to each item in Observable and
// returns a new Observable with applied items.
func (parent *Observable) Map(f interface{}) (o *Observable) {
	// check validation of f
	fv, ft := reflect.ValueOf(f), reflect.TypeOf(f)
	if fv.Kind() != reflect.Func {
		panic(ErrFuncFlip)
	}
	if ft.NumIn() != 1 && ft.NumOut() != 1 {
		panic(ErrFuncFlip)
	}

	o = parent.newTransformObservable("map")
	o.flip = fv.Interface()
	o.operator = mapflow
	return o
}

func mapflow(o *Observable) {
	fv := reflect.ValueOf(o.flip)
	o.connected = true

	in := o.pred.outflow
	var wg sync.WaitGroup
	var fn func(v interface{})

	// call map(v)
	switch fname := o.name; fname {
	case "map":
		fn = func(v interface{}) {
			defer wg.Done()

			var params = []reflect.Value{reflect.ValueOf(v)}
			res := fv.Call(params)
			if o.debug {
				fmt.Println(o.name, v, res[0])
			}
			o.outflow <- res[0].Interface()
		}
	case "filter":
		fn = func(v interface{}) {
			defer wg.Done()

			var params = []reflect.Value{reflect.ValueOf(v)}
			res := fv.Call(params)
			if o.debug {
				fmt.Println(o.name, v, res[0])
			}
			ok := res[0].Interface().(bool)
			if ok {
				o.outflow <- v
			}
		}
	}

	go func() {
		for x := range in {
			switch threading := o.threading; threading {
			case ThreadingDefault:
				wg.Add(1)
				fn(x)
			case ThreadingIO:
				fallthrough
			case ThreadingComputing:
				wg.Add(1)
				go fn(x)
			default:
			}
		}
		wg.Wait() //waiting all go-routines completed
		close(o.outflow)
	}()
}

// Filter `func(x anytype) bool` filters items in the original Observable and returns
// a new Observable with the filtered items.
func (parent *Observable) Filter(f interface{}) (o *Observable) {
	// check validation of f
	fv, ft := reflect.ValueOf(f), reflect.TypeOf(f)
	if fv.Kind() != reflect.Func {
		panic(ErrFuncFlip)
	}
	if ft.NumIn() != 1 && ft.NumOut() != 1 && ft.Out(0).Kind() != reflect.Bool {
		panic(ErrFuncFlip)
	}

	o = parent.newTransformObservable("filter")
	o.flip = fv.Interface()
	o.operator = mapflow
	return o
}

func (parent *Observable) newTransformObservable(name string) (o *Observable) {
	//new Observable
	o = newObservable()
	o.name = name

	//chain Observables
	parent.next = o
	o.pred = parent
	o.root = parent.root

	//set operators
	o.outflow = make(chan interface{}, BufferLen)
	return o
}
