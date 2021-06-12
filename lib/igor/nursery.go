package igor

import (
	"context"
	"sync"
)

type Nursery interface {
	// Close cancels all the goroutines in the nursery.
	Close(context.Context)
	// Wait waits while all the goroutinges in the nursery finish. If one of the goroutines in the
	// nursery failed with an error, that error is returned to the caller.
	Wait(context.Context) error
	// Checks whether nursery failed. If so, error is retured. If not so, nil is returned.
	// This function can be used for periodic checking of the health of the nursery.
	Err(context.Context) error
	// Methogs used internally by Igor.
	Context__() context.Context
	Start__()
	Stop__(error)
}

func NewNursery() Nursery {
	n := &nursery{}
	n.context, n.cancel = context.WithCancel(context.Background())
	return n
}

type nursery struct {
	context   context.Context
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup
	// First error reported by any goroutine in the nursery.
	// All subsequent errors are ignored.
	err error
}

func (n *nursery) Close(context.Context) {
	// Close should be as resilient as possible, so let's not fail here.
	if n == nil {
		return
	}
	n.cancel()
	n.Wait(nil)
}

func (n *nursery) Wait(context.Context) error {
	if n == nil {
		panic("nil nursery")
	}
	n.waitGroup.Wait()
	return n.err
}

func (n *nursery) Err(context.Context) error {
	if n == nil {
		panic("nil nursery")
	}
	return n.err
}

func (n *nursery) Context__() context.Context {
	if n == nil {
		panic("nil nursery")
	}
	return n.context
}

func (n *nursery) Start__() {
	if n == nil {
		panic("nil nursery")
	}
	n.waitGroup.Add(1)
}

func (n *nursery) Stop__(err error) {
	if n == nil {
		panic("nil nursery")
	}
	if n.err == nil {
		n.err = err
		if err != nil {
			// Cancel all the sibling goroutines.
			n.cancel()
		}
	}
	n.waitGroup.Done()
}
