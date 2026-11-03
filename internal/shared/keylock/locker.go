// Package keylock provides cancellable, process-local keyed exclusion.
package keylock

import (
	"context"
	"sync"
)

type entry struct {
	token chan struct{}
	refs  int
}

// Locker must be shared by all operations on the same resource in a process.
// Its zero value is ready to use. A Locker must not be copied after first use.
type Locker struct {
	mu      sync.Mutex
	entries map[string]*entry
}

func New() *Locker { return &Locker{} }

func (l *Locker) WithLock(ctx context.Context, key string, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.mu.Lock()
	if l.entries == nil {
		l.entries = make(map[string]*entry)
	}
	e := l.entries[key]
	if e == nil {
		e = &entry{token: make(chan struct{}, 1)}
		e.token <- struct{}{}
		l.entries[key] = e
	}
	e.refs++
	l.mu.Unlock()
	defer func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		e.refs--
		if e.refs == 0 {
			delete(l.entries, key)
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.token:
	}
	defer func() { e.token <- struct{}{} }()
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn()
}
