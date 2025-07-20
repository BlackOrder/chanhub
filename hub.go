package chanhub

import (
	"context"
	"sync"
)

func New() *Hub {
	return &Hub{subs: make(map[chan struct{}]struct{})}
}

type Hub struct {
	mu   sync.RWMutex
	subs map[chan struct{}]struct{}
}

func (h *Hub) Subscribe(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	go func() {
		<-ctx.Done()
		h.mu.Lock()
		delete(h.subs, ch)
		h.mu.Unlock()
		close(ch)
	}()
	return ch
}

func (h *Hub) Broadcast() {
	h.mu.RLock()
	for ch := range h.subs {
		select {
		case ch <- struct{}{}: // non-blocking
		default: // previous signal still pending – skip
		}
	}
	h.mu.RUnlock()
}
