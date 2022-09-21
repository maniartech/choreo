package conductor_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	conductor "github.com/maniartech/async"
	"github.com/stretchr/testify/assert"
)

func TestGoFutureBase(t *testing.T) {
	future := conductor.Func(processAsync, "A", 1000)

	isFuture := false

	if _, ok := interface{}(future).(*conductor.Future); ok {
		isFuture = true
	}

	assert.Equal(t, true, isFuture)
	assert.Equal(t, true, future.NotStarted())
	assert.Equal(t, false, future.Pending())
	assert.Equal(t, false, future.Finished())
}

func TestGoFuture(t *testing.T) {
	future := conductor.Func(processAsync, "A", 1000)
	result, err := future.Await()

	assert.Equal(t, true, future.Finished())

	assert.Equal(t, "A", result)
	assert.Equal(t, nil, err)

	future = conductor.Func(processAsync, "A", 1000, errors.New("invalid-action"))
	result, err = future.Await()

	assert.Equal(t, true, future.Finished())

	assert.Equal(t, nil, result)
	assert.EqualError(t, err, "invalid-action")

	_, err = future.Futures()
	assert.Error(t, err, "not-a-batch")
}

func TestBatchGo(t *testing.T) {
	vals := make([]string, 0)
	newCB := func() func(string) {
		return func(s string) {
			vals = append(vals, s)
		}
	}

	p := conductor.Async(
		conductor.Func(processAsync, "A", 3000, newCB()),
		conductor.Func(processAsync, "B", 2000, newCB()),
		conductor.Sync( // Calls Func routines in queue!
			conductor.Func(processAsync, "C", 1000, newCB()),
			conductor.Func(processAsync, "D", 500, newCB()),
			conductor.Func(processAsync, "E", 100, newCB()),
		),
		conductor.Async(
			conductor.Func(processAsync, "F", 200, newCB()),
			conductor.Func(processAsync, "G", 0, newCB()),
		),
	)

	assert.Equal(t, true, p.NotStarted())
	p.Await()
	childFutures, err := p.Futures()

	assert.Equal(t, true, p.Finished())
	assert.Equal(t, true, err == nil)
	assert.Equal(t, 4, len(childFutures))
	assert.Equal(t, "G,F,C,D,E,B,A", strings.Join(vals, ","))
}

func processAsync(p *conductor.Future, args ...interface{}) {
	s := args[0].(string)
	ms := args[1].(int)

	time.Sleep(time.Duration(ms) * time.Millisecond)

	defer func() {
		// If callback is supplied, call it by passing s!
		if len(args) == 3 {
			switch args[2].(type) {
			case func(string):
				p.Done(s)
				cb := args[2].(func(string))
				cb(s)
			case error:
				p.Done(args[2])
			default:
				p.Done(s)
			}
			return
		}
		p.Done(s)
	}()
}
