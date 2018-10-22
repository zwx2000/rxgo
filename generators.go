// Copyright 2018 The SS.SYSU Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rxgo

import (
	"context"
	"reflect"
)

// start flip which channel is closed by itself
func generator(ctx context.Context, o *Observable) {
	fv := reflect.ValueOf(o.flip)
	params := []reflect.Value{reflect.ValueOf(ctx)}

	go fv.Call(params)
}

// start flip as func() (x anytype, end bool)
func generatorCustomFunc(ctx context.Context, o *Observable) {
	fv := reflect.ValueOf(o.flip)
	params := []reflect.Value{}
	if o.flip_sup_ctx {
		params = []reflect.Value{reflect.ValueOf(ctx)}
	}

	go func() {
		for end := false; !end; {
			rs, skip, stop, e := userFuncCall(fv, params)
			end, _ = (rs[1].Interface()).(bool)
			var item interface{} = rs[0].Interface()
			if stop {
				end = true
			}
			if skip {
				continue
			}
			if e != nil {
				item = e
			}

			if !end {
				select {
				case o.outflow <- item:
					if o.debug != nil {
						o.debug.OnNext(item)
					}
				case <-ctx.Done():
					end = true
				}
			}

		}

		o.closeFlow()
	}()
}

// Generator creates an Observable with the provided item(s) producing by the function `func()  (val anytype, end bool)`
func Generator(name string, f interface{}) *Observable {
	fv := reflect.ValueOf(f)
	inType := []reflect.Type{}
	outType := []reflect.Type{typeAny, typeBool}
	ctx_sup := false
	if b, cb := checkFuncUpcast(fv, inType, outType, true); !b {
		panic(ErrFuncFlip)
	} else {
		ctx_sup = cb
	}

	o := newGeneratorObservable(name + "Generator")
	o.flip_sup_ctx = ctx_sup

	o.flip = fv.Interface()
	o.operator = generatorCustomFunc
	return o
}

// Generator creates an Observable with the provided item(s) producing by the function `func()  (val anytype, end bool)`
func Start(f interface{}) *Observable {
	return Generator("Start", f)
}

// Range creates an Observable that emits a particular range of sequential integers.
func Range(start, end int) *Observable {
	o := newGeneratorObservable("Range")

	o.flip = func(ctx context.Context) {
		i := start
		for i < end {
			select {
			case o.outflow <- i:
				if o.debug != nil {
					o.debug.OnNext(i)
				}
			case <-ctx.Done():
				goto outfor
			}
			i++
		}
	outfor:
		o.closeFlow()
	}
	o.operator = generator
	return o
}

// Just creates an Observable with the provided item(s).
func Just(items ...interface{}) *Observable {
	o := newGeneratorObservable("Just")

	o.flip = func(ctx context.Context) {
		for _, item := range items {
			select {
			case o.outflow <- item:
				if o.debug != nil {
					o.debug.OnNext(item)
				}
			case <-ctx.Done():
				goto outfor
			}
		}
	outfor:
		o.closeFlow()
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

		o.flip = func(ctx context.Context) {
			i := 0
			for i < length {
				select {
				case o.outflow <- v.Index(i).Interface():
					if o.debug != nil {
						o.debug.OnNext(v.Index(i).Interface())
					}
				case <-ctx.Done():
					goto outfor
				}
				i++
			}
		outfor:
			o.closeFlow()
		}
		o.operator = generator
		return o
	}

	if v.Kind() == reflect.Chan {
		o := newGeneratorObservable("From Channel")

		o.flip = func(ctx context.Context) {
			for {
				val, ok := v.Recv()
				if !ok {
					break
				}
				select {
				case o.outflow <- val.Interface():
					if o.debug != nil {
						o.debug.OnNext(val.Interface())
					}
				case <-ctx.Done():
					goto outfor
				}
			}
		outfor:
			o.closeFlow()
		}
		o.operator = generator
		return o
	}

	st := reflect.TypeOf((*Observable)(nil))
	//fmt.Println(t, st)
	if t == st {
		o := newGeneratorObservable("From *Observable")

		o.flip = func(ctx context.Context) {
			ro := v.Interface().(*Observable)
			ro.connect(ctx)
			for ; ro.next != nil; ro = ro.next {
			}
			ch := ro.outflow
			for x := range ch {
				select {
				case o.outflow <- x:
				case <-ctx.Done():
					goto outfor
				}
			}
		outfor:
			o.closeFlow()
		}
		o.operator = generator
		return o
	}

	panic(ErrFuncFlip)
}

// create an Observable that emits no items and does not terminate.
// It is important for combining with other Observables
func Never() *Observable {
	o := newGeneratorObservable("Never")

	o.flip = func(ctx context.Context) {
		select {
		case <-ctx.Done():
		}
		o.closeFlow()
	}
	o.operator = generator
	return o
}

// create an Observable that emits no items but terminates normally
func Empty() *Observable {
	o := newGeneratorObservable("Empty")

	o.flip = func(ctx context.Context) {
		o.closeFlow()
	}
	o.operator = generator
	return o
}

// create an Observable that emits no items and terminates with an error
func Throw(e error) *Observable {
	o := newGeneratorObservable("Throw")

	o.flip = func(ctx context.Context) {
		select {
		case o.outflow <- e:
		case <-ctx.Done():
		}
		o.closeFlow()
	}
	o.operator = generator
	return o
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
