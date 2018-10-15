// Copyright 2018 The SS.SYSU Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rxgo

import (
	"errors"
	"reflect"
)

var EoFlow = errors.New("End of Flow!")

// start flip which channel is closed by itself
func generator(o *Observable) {
	fv := reflect.ValueOf(o.flip)
	params := []reflect.Value{}
	o.connected = true
	go fv.Call(params)
}

// start flip as func() (x anytype, end bool)
func generatorCustomFunc(o *Observable) {
	fv := reflect.ValueOf(o.flip)
	params := []reflect.Value{}
	o.connected = true

	go func() {
		end := false
		for !end {
			rs := fv.Call(params)
			end, _ = (rs[1].Interface()).(bool)
			if !end {
				o.outflow <- rs[0].Interface()
			}
		}
		close(o.outflow)
	}()
}

// Generator creates an Observable with the provided item(s) producing by the function `func()  (val anytype, end bool)`
func Generator(f interface{}) *Observable {
	fv, ft := reflect.ValueOf(f), reflect.TypeOf(f)
	if fv.Kind() != reflect.Func {
		panic(ErrFuncFlip)
	}
	if ft.NumIn() == 0 && ft.NumOut() != 2 && ft.Out(1).Kind() != reflect.Bool {
		panic(ErrFuncFlip)
	}

	o := newGeneratorObservable("Custom Generator")

	o.flip = fv.Interface()
	o.operator = generatorCustomFunc
	return o
}

func Start(f interface{}) *Observable {
	return Generator(f)
}

// Range creates an Observable that emits a particular range of sequential integers.
func Range(start, end int) *Observable {
	o := newGeneratorObservable("Range")

	o.flip = func() {
		i := start
		for i < end {
			o.outflow <- i
			i++
		}
		close(o.outflow)
	}
	o.operator = generator
	return o
}

// Just creates an Observable with the provided item(s).
func Just(items ...interface{}) *Observable {
	o := newGeneratorObservable("Just")

	o.flip = func() {
		for _, item := range items {
			o.outflow <- item
		}
		close(o.outflow)
	}
	o.operator = generator
	return o
}

// convert Slice, Channel, and Observable into Observables
func From(items interface{}) *Observable {
	v, t := reflect.ValueOf(items), reflect.TypeOf(items)

	if v.Kind() == reflect.Slice {
		length := v.Len()
		o := newGeneratorObservable("From Slice")

		o.flip = func() {
			i := 0
			for i < length {
				o.outflow <- v.Index(i).Interface()
				i++
			}
			close(o.outflow)
		}
		o.operator = generator
		return o
	}

	if v.Kind() == reflect.Chan {
		o := newGeneratorObservable("From Channel")

		o.flip = func() {
			for {
				val, ok := v.Recv()
				if !ok {
					break
				}
				o.outflow <- val.Interface()
			}
			close(o.outflow)
		}
		o.operator = generator
		return o
	}

	st := reflect.TypeOf((*Observable)(nil))
	//fmt.Println(t, st)
	if t == st {
		o := newGeneratorObservable("From *Observable")

		o.flip = func() {
			ro := v.Interface().(*Observable)
			ro.connect()
			for ; ro.next != nil; ro = ro.next {
			}
			ch := ro.outflow
			for x := range ch {
				o.outflow <- x
			}
			close(o.outflow)
		}
		o.operator = generator
		return o
	}

	panic(ErrFuncFlip)
}

func newGeneratorObservable(name string) (o *Observable) {
	//new Observable
	o = newObservable()
	o.name = name

	//chain Observables
	o.root = o

	//set options
	o.buf_len = 0
	return o
}
