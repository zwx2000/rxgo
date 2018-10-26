package rxgo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pmlpml/rxgo"
	"github.com/stretchr/testify/assert"
)

func TestAnyTranform(t *testing.T) {
	iCount := 0
	sCount := 0
	eCount := 0
	res := []int{}

	rxgo.Generator(func(ctx context.Context, send func(x interface{}) (endSignal bool)) {
		send(10)
		send("hello")
		send(20)
		send(errors.New("Any"))
		send(30)
	}).TransformOp(func(ctx context.Context, item interface{}, send func(x interface{}) (endSignal bool)) {
		if i, ok := item.(int); ok {
			send(i + 1)
		} else {
			send(item)
		}
	}).Subscribe(rxgo.ObserverMonitor{
		Next: func(item interface{}) {
			if i, ok := item.(int); ok {
				iCount++
				res = append(res, i)
			} else {
				sCount++
			}
		},
		Error: func(e error) {
			eCount++
		},
	})

	assert.Equal(t, []int{3, 1, 1}, []int{iCount, sCount, eCount}, "type count error")
	assert.Equal(t, []int{11, 21, 31}, res, "transform data error")
}
