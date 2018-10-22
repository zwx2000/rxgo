// Copyright 2018 The SS.SYSU Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rxgo

import "fmt"

// Test Observer
type InnerObserver struct {
	name string
}

var _ Observer = InnerObserver{"test"}

func (o InnerObserver) OnNext(x interface{}) {
	fmt.Println(o.name, "Receive value ", x)
}

func (o InnerObserver) OnError(e error) {
	fmt.Println(o.name, "Error ", e)
}

func (o InnerObserver) OnCompleted() {
	fmt.Println(o.name, "Down ")
}
