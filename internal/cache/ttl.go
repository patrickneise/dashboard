package cache

import (
	"sync"
	"time"
)

type State uint8

const (
	Miss State = iota
	Fresh
	Stale
)

type TTL[T any] struct {
	mu  sync.Mutex
	v   T
	ok  bool
	exp time.Time
}

func (c *TTL[T]) Get(now time.Time) (T, time.Time, State) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var zero T
	if !c.ok {
		return zero, time.Time{}, Miss
	}

	// Keep value even if expired (stale)
	if now.After(c.exp) {
		return c.v, c.exp, Stale
	}

	return c.v, c.exp, Fresh
}

func (c *TTL[T]) Set(v T, exp time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.v = v
	c.exp = exp
	c.ok = true
}
