// Package gate provides primitive to limit number of concurrent goroutine
// workers. Useful when sync.Locker or sync.WaitGroup is not enough.
package gate

import (
	"runtime"
	"sync"
)

// A Gate is a primitive intended to help in limiting concurrency in some
// scenarios. Think of it as a close sync.WaitGroup analog which has upper
// limit on its counter or a sync.Locker which allows up to max number of
// concurrent lockers to be held.
type Gate struct {
	c chan struct{}
	m sync.Mutex
}

var defaultGate = New(runtime.NumCPU())

// Lock locks default gate with capacity defined by runtime.NumCPU()
func Lock() { defaultGate.Lock() }

// Unlock unlocks default gate
func Unlock() { defaultGate.Unlock() }

// Add adds n to default gate counter. Absolute value of n should be no more
// than runtime.NumCPU()
func Add(n int) { defaultGate.Add(n) }

// Done decrements default gate counter
func Done() { defaultGate.Done() }

// Wait blocks until default gate is not locked
func Wait() { defaultGate.Wait() }

// New returns new Gate with provided capacity. If capacity is non-positive,
// New would panic.
func New(max int) *Gate { return &Gate{c: make(chan struct{}, max)} }

// Lock implements sync.Locker interface. Gate capacity determines number of
// non-blocking Lock calls, when max number is reached, Lock would block until
// some other goroutine calls Unlock. Lock is safe for concurrent use.
func (g *Gate) Lock() { g.c <- struct{}{} }

// Unlock implements sync.Locker interface. Unlock is safe for concurrent use.
func (g *Gate) Unlock() { <-g.c }

// Add implements similar semantic to sync.WaitGroup.Add. If Add is called with
// positive argument N, it essentially calls Lock N times; if N is negative, it
// calls Unlock N times. If absolute value of N is greater than Gate capacity,
// Add would panic. Add is safe for concurrent use, but should be used with
// care as deadlocks are possible.
func (g *Gate) Add(n int) {
	if c := cap(g.c); n > c || -n > c {
		panic("gate: out of range Add argument")
	}
	if n > 0 {
		for i := 0; i < n; i++ {
			g.c <- struct{}{}
		}
		return
	}
	for i := 0; i < (-n); i++ {
		<-g.c
	}
}

// Done semantic is the same as sync.WaitGroup.Done.
func (g *Gate) Done() { <-g.c }

// Wait blocks until nothing holds a single Gate lock. Its semantic is the same
// as sync.WaitGroup.Wait.
func (g *Gate) Wait() {
	g.m.Lock()
	defer g.m.Unlock()
	for i := 0; i < cap(g.c); i++ {
		g.c <- struct{}{}
	}
	for i := 0; i < cap(g.c); i++ {
		<-g.c
	}
}
