// Copyright 2018 The SS.SYSU Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rxgo

import (
	"context"
	"reflect"
	"sync"
)

var (
	typeAny        = reflect.TypeOf((*interface{})(nil)).Elem()
	typeContext    = reflect.TypeOf((*context.Context)(nil)).Elem()
	typeBool       = reflect.TypeOf(true)
	typeObservable = reflect.TypeOf(&Observable{})
)

// func type check, such as `func(x int) bool` satisfy `func(x anytype) bool`
func checkFuncUpcast(fv reflect.Value, inType, outType []reflect.Type, ctx_sup bool) (b, ctx_b bool) {
	//fmt.Println(fv.Kind(),reflect.Func)
	if fv.Kind() != reflect.Func {
		return // Not func
	}
	ft := fv.Type()
	if ft.NumOut() != len(outType) {
		return // Error result parameters
	}
	if !ctx_sup {
		if ft.NumIn() != len(inType) {
			return
		}
	} else {
		if ft.NumIn() == 0 {
			if len(inType) != 0 {
				return
			}
		} else {
			if ft.In(0).Implements(typeContext) {
				ctx_b = true
				if ft.NumIn() != len(inType)+1 {
					return
				}
			} else {
				if ft.NumIn() != len(inType) {
					return
				}
			}
		}
	}

	for i, t := range inType {
		var real_t reflect.Type
		if ctx_b {
			real_t = ft.In(i + 1)
		} else {
			real_t = ft.In(i)
		}

		//todo: ptr or slice check
		switch {
		case real_t == t:
		case t.Kind() == reflect.Interface && real_t.Implements(t):
		//case ft.In(i).AssignableTo(t):
		//case ft.In(i).ConvertibleTo(t):
		default:
			return
		}
	}
	for i, t := range outType {
		//fmt.Println(ft.Out(i), t)
		//todo: ptr or slice check
		switch {
		case ft.Out(i) == t:
		case t.Kind() == reflect.Interface && ft.Out(i).Implements(t):
		default:
			return
		}
	}
	b = true
	return
}

// wrap exception when call user function
func userFuncCall(fv reflect.Value, params []reflect.Value) (res []reflect.Value, skip, stop bool, eout error) {
	defer func() {
		if e := recover(); e != nil {
			if fe, ok := e.(FlowableError); ok {
				eout = fe
			}
			switch e {
			case ErrSkipItem:
				skip = true
				return
			case ErrEoFlow:
				stop = true
				return
			default:
				panic(e)
			}
		}
	}()

	res = fv.Call(params)
	return
}

// Map maps each item in Observable by the function with `func(x anytype) anytype` and
// returns a new Observable with applied items.
func (parent *Observable) Map(f interface{}) (o *Observable) {
	// check validation of f
	fv := reflect.ValueOf(f)
	inType := []reflect.Type{typeAny}
	outType := []reflect.Type{typeAny}
	if b, _ := checkFuncUpcast(fv, inType, outType, false); !b {
		panic(ErrFuncFlip)
	}

	o = parent.newTransformObservable("map")
	o.flip = fv.Interface()
	o.operator = mapflow
	return o
}

// FlatMap maps each item in Observable by the function with `func(x anytype) (o *Observable) ` and
// returns a new Observable with merged observables appling on each items.
func (parent *Observable) FlatMap(f interface{}) (o *Observable) {
	// check validation of f
	fv := reflect.ValueOf(f)
	inType := []reflect.Type{typeAny}
	outType := []reflect.Type{typeObservable}
	if b, _ := checkFuncUpcast(fv, inType, outType, false); !b {
		panic(ErrFuncFlip)
	}

	o = parent.newTransformObservable("flatMap")
	o.flip = fv.Interface()
	o.operator = mapflow
	return o
}

func mapflow(ctx context.Context, o *Observable) {
	fv := reflect.ValueOf(o.flip)

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
			if o.debug != nil {
				o.debug.OnNext(res[0].Interface())
			}
			o.outflow <- res[0].Interface()
		}
	case "flatMap":
		fn = func(v interface{}) {
			defer wg.Done()

			var params = []reflect.Value{reflect.ValueOf(v)}
			res := fv.Call(params)

			ro := res[0].Interface().(*Observable)
			if ro != nil {
				// subscribe ro without any ObserveOn model
				ro.connect(ctx)
				for ; ro.next != nil; ro = ro.next {
				}
				ch := ro.outflow
				for x := range ch {
					o.outflow <- x
					if o.debug != nil {
						o.debug.OnNext(x)
					}
				}
			}
		}
	case "filter":
		fn = func(v interface{}) {
			defer wg.Done()

			var params = []reflect.Value{reflect.ValueOf(v)}
			res := fv.Call(params)

			ok := res[0].Interface().(bool)
			if ok {
				o.outflow <- v
				if o.debug != nil {
					o.debug.OnNext(v)
				}
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
		o.closeFlow()
	}()
}

// Filter `func(x anytype) bool` filters items in the original Observable and returns
// a new Observable with the filtered items.
func (parent *Observable) Filter(f interface{}) (o *Observable) {
	// check validation of f
	fv := reflect.ValueOf(f)
	inType := []reflect.Type{typeAny}
	outType := []reflect.Type{typeBool}
	if b, _ := checkFuncUpcast(fv, inType, outType, false); !b {
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

	//set options
	o.buf_len = BufferLen
	return o
}
