package choreo

import (
	"fmt"
	"sync"
)

// startAsync starts the specified futures in parallel go routines.
func startAsync(p *Choreography, args ...interface{}) {
	wg := sync.WaitGroup{}

	for i := 0; i < len(args); i++ {
		pr, ok := args[i].(*Choreography)
		if !ok {
			panic(fmt.Errorf("%s at '%v'", errInvalidChoreography, i))
		} else if pr.Pending() {
			panic(fmt.Errorf("%s at '%v'", errInvalidState, i))
		}
		wg.Add(1)
		pr.Then(func(interface{}, error) {
			wg.Done()
		})
		pr.Start()
	}
	wg.Wait()
	p.Done()
}

// startSync starts the specified futures in new go routines with queued mannger.
// That is it starts the future only after the preview future has finished.
func startSync(p *Choreography, args ...interface{}) {
	for i := 0; i < len(args); i++ {
		pr, ok := args[i].(*Choreography)
		if !ok {
			panic(fmt.Errorf("%s at '%v'", errInvalidChoreography, i))
		} else if pr.Pending() {
			panic(fmt.Errorf("%s at '%v'", errInvalidState, i))
		}
		pr.Await()
	}
	p.Done()
}

// Create creates a future task runner that executes single task.
func create(fn ChoreographyHandler, args ...interface{}) *Choreography {
	return &Choreography{
		fn:   fn,
		args: args,
		wg:   sync.WaitGroup{},
	}
}

// Creates a future that executes one or more handlers
func createBatch(q bool, futures ...*Choreography) *Choreography {
	if len(futures) == 0 {
		panic(errInvalidArguments)
	}

	interfaces := make([]interface{}, len(futures))
	for i := 0; i < len(futures); i++ {
		interfaces[i] = futures[i]
	}

	var p *Choreography

	if q {
		p = create(startSync, interfaces...)
	} else {
		p = create(startAsync, interfaces...)
	}
	p.batch = true
	return p
}
