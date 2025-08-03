/*
Package chanhub provides a lightweight, thread-safe hub for broadcasting signals
to multiple subscribers using channels. It's designed for implementing pub/sub
patterns, event notification systems, and coordinating goroutines.

# Features

• Thread-safe: All operations are protected by read-write mutexes
• Context-aware: Automatic cleanup when contexts are canceled
• Non-blocking broadcasts: Broadcasts never block, even if subscribers are slow
• Buffered channels: Each subscriber gets a buffered channel to prevent blocking
• Automatic cleanup: Subscribers are automatically removed when their context is canceled
• Zero dependencies: Uses only the Go standard library
• High performance: Optimized for concurrent access and minimal overhead

# Basic Usage

	hub := chanhub.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe to the hub
	ch := hub.Subscribe(ctx)

	// Listen for signals in a goroutine
	go func() {
		for {
			select {
			case <-ch:
				fmt.Println("Signal received!")
			case <-ctx.Done():
				return
			}
		}
	}()

	// Broadcast a signal to all subscribers
	hub.Broadcast()

# Use Cases

Event Notification Systems:
Perfect for notifying multiple components when events occur, such as configuration
changes, user actions, or system state changes.

Worker Coordination:
Coordinate the start and stop of multiple goroutines, ensuring synchronized
execution across your application.

Configuration Reloading:
Notify all services when configuration files change, enabling live reloading
without restarts.

Cache Invalidation:
Broadcast cache invalidation signals to multiple cache instances in a
distributed system.

Real-time Updates:
Implement server-sent events or WebSocket broadcasting for real-time web
applications.

# Thread Safety

All methods of Hub are safe for concurrent use. Multiple goroutines can
simultaneously:
• Subscribe to the hub
• Broadcast signals
• Cancel contexts (triggering automatic cleanup)

The internal state is protected by a read-write mutex, ensuring data consistency
while allowing concurrent reads during broadcasts.

# Performance Characteristics

• Subscribe: O(1) operation
• Broadcast: O(n) where n is the number of subscribers, but non-blocking
• Memory overhead: Minimal - only channel references stored in a map
• Cleanup: Automatic with no manual intervention required

# Context Handling

Subscribers are automatically cleaned up when their associated context is
canceled. This prevents memory leaks and ensures that canceled goroutines
don't continue to receive signals.

The Subscribe method accepts a context.Context parameter. When this context
is canceled:
1. The subscription is removed from the hub
2. The subscriber's channel is closed
3. The cleanup goroutine terminates

This design ensures that long-running applications can safely create and
destroy subscriptions without memory leaks.
*/
package chanhub
