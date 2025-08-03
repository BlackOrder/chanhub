package chanhub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	hub := New()
	if hub == nil {
		t.Fatal("New() returned nil")
	}
	if hub.subs == nil {
		t.Fatal("New() did not initialize subs map")
	}
	if len(hub.subs) != 0 {
		t.Fatalf("New() should start with empty subs map, got %d entries", len(hub.subs))
	}
}

func TestHub_Subscribe(t *testing.T) {
	hub := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)
	if ch == nil {
		t.Fatal("Subscribe() returned nil channel")
	}

	// Check that subscription was added
	hub.mu.RLock()
	subsCount := len(hub.subs)
	hub.mu.RUnlock()

	if subsCount != 1 {
		t.Fatalf("Expected 1 subscription, got %d", subsCount)
	}
}

func TestHub_SubscribeContextCancellation(t *testing.T) {
	hub := New()
	ctx, cancel := context.WithCancel(context.Background())

	ch := hub.Subscribe(ctx)

	// Verify subscription exists
	hub.mu.RLock()
	initialCount := len(hub.subs)
	hub.mu.RUnlock()

	if initialCount != 1 {
		t.Fatalf("Expected 1 subscription, got %d", initialCount)
	}

	// Cancel context
	cancel()

	// Wait for cleanup
	time.Sleep(10 * time.Millisecond)

	// Verify subscription was removed
	hub.mu.RLock()
	finalCount := len(hub.subs)
	hub.mu.RUnlock()

	if finalCount != 0 {
		t.Fatalf("Expected 0 subscriptions after context cancellation, got %d", finalCount)
	}

	// Verify channel is closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("Channel should be closed after context cancellation")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel was not closed within timeout")
	}
}

func TestHub_MultipleSubscriptions(t *testing.T) {
	hub := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const numSubs = 5
	channels := make([]<-chan struct{}, numSubs)

	// Create multiple subscriptions
	for i := 0; i < numSubs; i++ {
		channels[i] = hub.Subscribe(ctx)
	}

	// Verify all subscriptions exist
	hub.mu.RLock()
	if len(hub.subs) != numSubs {
		t.Fatalf("Expected %d subscriptions, got %d", numSubs, len(hub.subs))
	}
	hub.mu.RUnlock()

	// Verify all channels are unique
	channelSet := make(map[<-chan struct{}]bool)
	for _, ch := range channels {
		if channelSet[ch] {
			t.Fatal("Duplicate channel returned by Subscribe()")
		}
		channelSet[ch] = true
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const numSubs = 3
	channels := make([]<-chan struct{}, numSubs)

	// Create subscriptions
	for i := 0; i < numSubs; i++ {
		channels[i] = hub.Subscribe(ctx)
	}

	// Broadcast signal
	hub.Broadcast()

	// Verify all channels received signal
	for i, ch := range channels {
		select {
		case <-ch:
			// Signal received successfully
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Channel %d did not receive signal within timeout", i)
		}
	}
}

func TestHub_BroadcastNoSubscribers(t *testing.T) {
	hub := New()

	// This should not panic or cause issues
	hub.Broadcast()

	// Verify no subscriptions exist
	hub.mu.RLock()
	if len(hub.subs) != 0 {
		t.Fatalf("Expected 0 subscriptions, got %d", len(hub.subs))
	}
	hub.mu.RUnlock()
}

func TestHub_BroadcastNonBlocking(t *testing.T) {
	hub := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	// First broadcast should send signal
	hub.Broadcast()

	// Verify signal is in channel
	select {
	case <-ch:
		// Good, received first signal
	case <-time.After(100 * time.Millisecond):
		t.Fatal("First broadcast signal not received")
	}

	// Second broadcast should not block even though channel buffer is empty
	// but previous signal hasn't been consumed
	done := make(chan bool)
	go func() {
		hub.Broadcast()
		done <- true
	}()

	select {
	case <-done:
		// Good, broadcast didn't block
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Second broadcast blocked unexpectedly")
	}
}

func TestHub_ConcurrentSubscribeAndBroadcast(t *testing.T) {
	hub := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const numGoroutines = 10
	const numBroadcasts = 5

	var wg sync.WaitGroup
	receivedSignals := make([]int, numGoroutines)

	// Start subscriber goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ch := hub.Subscribe(ctx)

			for {
				select {
				case <-ch:
					receivedSignals[idx]++
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	// Give time for subscriptions to be established
	time.Sleep(10 * time.Millisecond)

	// Send broadcasts
	for i := 0; i < numBroadcasts; i++ {
		hub.Broadcast()
		time.Sleep(5 * time.Millisecond)
	}

	// Wait a bit for signals to be processed
	time.Sleep(50 * time.Millisecond)

	// Cancel context to stop goroutines
	cancel()
	wg.Wait()

	// Verify all goroutines received at least some signals
	for i, count := range receivedSignals {
		if count == 0 {
			t.Errorf("Goroutine %d received no signals", i)
		}
	}
}

func TestHub_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	hub := New()

	const numGoroutines = 10
	var wg sync.WaitGroup

	// Start goroutines that subscribe and immediately cancel
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			hub.Subscribe(ctx)
			time.Sleep(time.Millisecond) // Brief delay
			cancel()
		}()
	}

	wg.Wait()

	// Wait for cleanup
	time.Sleep(50 * time.Millisecond)

	// Verify all subscriptions were cleaned up
	hub.mu.RLock()
	subsCount := len(hub.subs)
	hub.mu.RUnlock()

	if subsCount != 0 {
		t.Fatalf("Expected 0 subscriptions after cleanup, got %d", subsCount)
	}
}

func TestHub_ChannelBuffering(t *testing.T) {
	hub := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	// Send signal
	hub.Broadcast()

	// Verify signal is buffered and can be received later
	select {
	case <-ch:
		// Good, signal was buffered
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Buffered signal not received")
	}

	// Send another signal
	hub.Broadcast()

	// Verify second signal is also received
	select {
	case <-ch:
		// Good, second signal received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Second signal not received")
	}
}

// Benchmark tests
func BenchmarkHub_Subscribe(b *testing.B) {
	hub := New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.Subscribe(ctx)
	}
}

func BenchmarkHub_Broadcast(b *testing.B) {
	hub := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create some subscriptions
	const numSubs = 100
	for i := 0; i < numSubs; i++ {
		hub.Subscribe(ctx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.Broadcast()
	}
}

func BenchmarkHub_ConcurrentOperations(b *testing.B) {
	hub := New()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithCancel(context.Background())
			ch := hub.Subscribe(ctx)
			hub.Broadcast()
			select {
			case <-ch:
			default:
			}
			cancel()
		}
	})
}
