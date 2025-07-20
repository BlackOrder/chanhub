package chanhub

import (
	"context"
	"sync"
)

func New() *hub {
	return &hub{subs: make(map[chan struct{}]struct{})}
}

type hub struct {
	mu   sync.RWMutex
	subs map[chan struct{}]struct{}
}

func (h *hub) Subscribe(ctx context.Context) <-chan struct{} {
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

func (h *hub) Broadcast() {
	h.mu.RLock()
	for ch := range h.subs {
		select {
		case ch <- struct{}{}: // non-blocking
		default: // previous signal still pending – skip
		}
	}
	h.mu.RUnlock()
}
