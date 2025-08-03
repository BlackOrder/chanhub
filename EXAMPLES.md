# Examples

This document contains practical examples of using ChanHub in various scenarios.

## Table of Contents

1. [Basic Usage](#basic-usage)
2. [Web Server Notifications](#web-server-notifications)
3. [Database Connection Pool Events](#database-connection-pool-events)
4. [File Watcher Service](#file-watcher-service)
5. [Worker Pool Coordination](#worker-pool-coordination)
6. [Real-time Chat System](#real-time-chat-system)
7. [Cache Invalidation](#cache-invalidation)

## Basic Usage

### Simple Publisher-Subscriber

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/blackorder/chanhub"
)

func main() {
    hub := chanhub.New()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Create multiple subscribers
    subscriber1 := hub.Subscribe(ctx)
    subscriber2 := hub.Subscribe(ctx)
    subscriber3 := hub.Subscribe(ctx)

    // Start goroutines to listen
    go listen("Subscriber 1", subscriber1, ctx)
    go listen("Subscriber 2", subscriber2, ctx)
    go listen("Subscriber 3", subscriber3, ctx)

    // Send periodic broadcasts
    for i := 0; i < 5; i++ {
        fmt.Printf("Broadcasting message %d\n", i+1)
        hub.Broadcast()
        time.Sleep(1 * time.Second)
    }
}

func listen(name string, ch <-chan struct{}, ctx context.Context) {
    for {
        select {
        case <-ch:
            fmt.Printf("%s received signal\n", name)
        case <-ctx.Done():
            fmt.Printf("%s shutting down\n", name)
            return
        }
    }
}
```

## Web Server Notifications

### Server-Sent Events (SSE)

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/blackorder/chanhub"
)

type SSEServer struct {
    hub *chanhub.Hub
}

func NewSSEServer() *SSEServer {
    return &SSEServer{
        hub: chanhub.New(),
    }
}

func (s *SSEServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    ctx := r.Context()
    eventCh := s.hub.Subscribe(ctx)

    // Send initial connection event
    fmt.Fprintf(w, "data: Connected to event stream\n\n")
    w.(http.Flusher).Flush()

    for {
        select {
        case <-eventCh:
            // Send server-sent event
            fmt.Fprintf(w, "data: New event at %s\n\n", time.Now().Format(time.RFC3339))
            w.(http.Flusher).Flush()
        case <-ctx.Done():
            return
        }
    }
}

func (s *SSEServer) TriggerEvent() {
    s.hub.Broadcast()
}

func main() {
    server := NewSSEServer()

    // Handle SSE endpoint
    http.HandleFunc("/events", server.HandleSSE)

    // Handle trigger endpoint
    http.HandleFunc("/trigger", func(w http.ResponseWriter, r *http.Request) {
        server.TriggerEvent()
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Event triggered"))
    })

    // Simulate periodic events
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        for range ticker.C {
            server.TriggerEvent()
        }
    }()

    fmt.Println("SSE Server running on :8080")
    fmt.Println("Open http://localhost:8080/events in your browser")
    fmt.Println("Trigger events at http://localhost:8080/trigger")
    
    http.ListenAndServe(":8080", nil)
}
```

## Database Connection Pool Events

### Connection Pool Status Notifications

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/blackorder/chanhub"
)

type ConnectionPool struct {
    mu          sync.RWMutex
    connections []string
    hub         *chanhub.Hub
    maxSize     int
}

func NewConnectionPool(maxSize int) *ConnectionPool {
    return &ConnectionPool{
        connections: make([]string, 0, maxSize),
        hub:         chanhub.New(),
        maxSize:     maxSize,
    }
}

func (cp *ConnectionPool) Subscribe(ctx context.Context) <-chan struct{} {
    return cp.hub.Subscribe(ctx)
}

func (cp *ConnectionPool) AddConnection(id string) {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    if len(cp.connections) < cp.maxSize {
        cp.connections = append(cp.connections, id)
        fmt.Printf("Added connection: %s (total: %d)\n", id, len(cp.connections))
        
        // Notify subscribers of pool change
        cp.hub.Broadcast()
    }
}

func (cp *ConnectionPool) RemoveConnection(id string) {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    for i, conn := range cp.connections {
        if conn == id {
            cp.connections = append(cp.connections[:i], cp.connections[i+1:]...)
            fmt.Printf("Removed connection: %s (total: %d)\n", id, len(cp.connections))
            
            // Notify subscribers of pool change
            cp.hub.Broadcast()
            break
        }
    }
}

func (cp *ConnectionPool) Status() (int, int) {
    cp.mu.RLock()
    defer cp.mu.RUnlock()
    return len(cp.connections), cp.maxSize
}

func main() {
    pool := NewConnectionPool(5)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Monitor pool changes
    poolChanges := pool.Subscribe(ctx)
    go func() {
        for {
            select {
            case <-poolChanges:
                current, max := pool.Status()
                usage := float64(current) / float64(max) * 100
                fmt.Printf("Pool status changed: %d/%d connections (%.1f%% usage)\n", 
                    current, max, usage)
                
                if usage > 80 {
                    fmt.Println("⚠️  High pool usage warning!")
                }
            case <-ctx.Done():
                fmt.Println("Pool monitor shutting down")
                return
            }
        }
    }()

    // Simulate connection management
    go func() {
        for i := 0; i < 10; i++ {
            time.Sleep(2 * time.Second)
            connID := fmt.Sprintf("conn_%d", i)
            pool.AddConnection(connID)
            
            // Remove some connections occasionally
            if i > 2 && i%3 == 0 {
                oldConnID := fmt.Sprintf("conn_%d", i-2)
                pool.RemoveConnection(oldConnID)
            }
        }
    }()

    time.Sleep(25 * time.Second)
}
```

## File Watcher Service

### Configuration File Monitoring

```go
package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/blackorder/chanhub"
)

type FileWatcher struct {
    filename string
    hub      *chanhub.Hub
    lastMod  time.Time
    stopCh   chan struct{}
}

func NewFileWatcher(filename string) *FileWatcher {
    return &FileWatcher{
        filename: filename,
        hub:      chanhub.New(),
        stopCh:   make(chan struct{}),
    }
}

func (fw *FileWatcher) Subscribe(ctx context.Context) <-chan struct{} {
    return fw.hub.Subscribe(ctx)
}

func (fw *FileWatcher) Start() error {
    // Get initial modification time
    info, err := os.Stat(fw.filename)
    if err != nil {
        return fmt.Errorf("failed to stat file: %w", err)
    }
    fw.lastMod = info.ModTime()

    go fw.watch()
    return nil
}

func (fw *FileWatcher) Stop() {
    close(fw.stopCh)
}

func (fw *FileWatcher) watch() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            info, err := os.Stat(fw.filename)
            if err != nil {
                fmt.Printf("Error checking file: %v\n", err)
                continue
            }

            if info.ModTime().After(fw.lastMod) {
                fw.lastMod = info.ModTime()
                fmt.Printf("File %s modified at %s\n", fw.filename, fw.lastMod.Format(time.RFC3339))
                fw.hub.Broadcast()
            }

        case <-fw.stopCh:
            return
        }
    }
}

func main() {
    // Create a test file
    testFile := "config.txt"
    if err := os.WriteFile(testFile, []byte("initial config"), 0644); err != nil {
        panic(err)
    }
    defer os.Remove(testFile)

    watcher := NewFileWatcher(testFile)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Subscribe to file changes
    fileChanges := watcher.Subscribe(ctx)

    // Start multiple services that react to config changes
    services := []string{"WebServer", "Database", "Cache", "Logger"}
    for _, service := range services {
        go func(serviceName string) {
            for {
                select {
                case <-fileChanges:
                    fmt.Printf("%s: Reloading configuration...\n", serviceName)
                    // Simulate config reload work
                    time.Sleep(100 * time.Millisecond)
                    fmt.Printf("%s: Configuration reloaded\n", serviceName)
                case <-ctx.Done():
                    fmt.Printf("%s: Shutting down\n", serviceName)
                    return
                }
            }
        }(service)
    }

    if err := watcher.Start(); err != nil {
        panic(err)
    }
    defer watcher.Stop()

    // Simulate file modifications
    go func() {
        for i := 0; i < 5; i++ {
            time.Sleep(3 * time.Second)
            content := fmt.Sprintf("updated config %d", i+1)
            if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
                fmt.Printf("Error writing file: %v\n", err)
            }
        }
    }()

    time.Sleep(20 * time.Second)
}
```

## Worker Pool Coordination

### Synchronized Start/Stop

```go
package main

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "time"

    "github.com/blackorder/chanhub"
)

type WorkerPool struct {
    startHub *chanhub.Hub
    stopHub  *chanhub.Hub
    workers  int
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        startHub: chanhub.New(),
        stopHub:  chanhub.New(),
        workers:  workers,
    }
}

func (wp *WorkerPool) StartAll(ctx context.Context) {
    var wg sync.WaitGroup

    // Start all workers
    for i := 0; i < wp.workers; i++ {
        wg.Add(1)
        go wp.worker(i, ctx, &wg)
    }

    // Give workers time to subscribe
    time.Sleep(100 * time.Millisecond)

    // Signal all workers to start
    fmt.Println("🚀 Starting all workers...")
    wp.startHub.Broadcast()

    wg.Wait()
    fmt.Println("✅ All workers completed")
}

func (wp *WorkerPool) StopAll() {
    fmt.Println("🛑 Stopping all workers...")
    wp.stopHub.Broadcast()
}

func (wp *WorkerPool) worker(id int, ctx context.Context, wg *sync.WaitGroup) {
    defer wg.Done()

    startCh := wp.startHub.Subscribe(ctx)
    stopCh := wp.stopHub.Subscribe(ctx)

    fmt.Printf("Worker %d: Ready and waiting for start signal\n", id)

    // Wait for start signal
    select {
    case <-startCh:
        fmt.Printf("Worker %d: Starting work\n", id)
    case <-ctx.Done():
        fmt.Printf("Worker %d: Context cancelled before start\n", id)
        return
    }

    // Simulate work with random duration
    workDuration := time.Duration(rand.Intn(5)+1) * time.Second
    workTimer := time.NewTimer(workDuration)
    defer workTimer.Stop()

    for {
        select {
        case <-workTimer.C:
            fmt.Printf("Worker %d: Work completed naturally\n", id)
            return

        case <-stopCh:
            fmt.Printf("Worker %d: Received stop signal, shutting down\n", id)
            return

        case <-ctx.Done():
            fmt.Printf("Worker %d: Context cancelled\n", id)
            return
        }
    }
}

func main() {
    pool := NewWorkerPool(5)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Start workers in a separate goroutine
    go pool.StartAll(ctx)

    // Simulate external stop command after some time
    go func() {
        time.Sleep(8 * time.Second)
        pool.StopAll()
    }()

    time.Sleep(15 * time.Second)
}
```

## Real-time Chat System

### Message Broadcast System

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"

    "github.com/blackorder/chanhub"
)

type ChatRoom struct {
    name    string
    hub     *chanhub.Hub
    mu      sync.RWMutex
    users   map[string]*User
    lastMsg string
}

type User struct {
    name string
    id   string
}

func NewChatRoom(name string) *ChatRoom {
    return &ChatRoom{
        name:  name,
        hub:   chanhub.New(),
        users: make(map[string]*User),
    }
}

func (cr *ChatRoom) Join(user *User, ctx context.Context) <-chan struct{} {
    cr.mu.Lock()
    cr.users[user.id] = user
    cr.mu.Unlock()

    fmt.Printf("📨 %s joined %s\n", user.name, cr.name)
    
    // Notify all users about new join
    cr.hub.Broadcast()
    
    return cr.hub.Subscribe(ctx)
}

func (cr *ChatRoom) Leave(userID string) {
    cr.mu.Lock()
    if user, exists := cr.users[userID]; exists {
        delete(cr.users, userID)
        fmt.Printf("👋 %s left %s\n", user.name, cr.name)
    }
    cr.mu.Unlock()

    // Notify remaining users
    cr.hub.Broadcast()
}

func (cr *ChatRoom) SendMessage(from *User, message string) {
    cr.mu.Lock()
    cr.lastMsg = fmt.Sprintf("%s: %s", from.name, message)
    cr.mu.Unlock()

    fmt.Printf("💬 [%s] %s\n", cr.name, cr.lastMsg)
    
    // Broadcast to all subscribers
    cr.hub.Broadcast()
}

func (cr *ChatRoom) GetInfo() (string, []string, string) {
    cr.mu.RLock()
    defer cr.mu.RUnlock()

    userNames := make([]string, 0, len(cr.users))
    for _, user := range cr.users {
        userNames = append(userNames, user.name)
    }

    return cr.name, userNames, cr.lastMsg
}

func main() {
    chatRoom := NewChatRoom("General")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Create some users
    users := []*User{
        {name: "Alice", id: "alice123"},
        {name: "Bob", id: "bob456"},
        {name: "Charlie", id: "charlie789"},
    }

    // Users join the chat room
    for _, user := range users {
        userCtx, userCancel := context.WithCancel(ctx)
        defer userCancel()

        msgCh := chatRoom.Join(user, userCtx)

        // Each user listens for updates
        go func(u *User, ch <-chan struct{}, cancel context.CancelFunc) {
            defer cancel()
            
            for {
                select {
                case <-ch:
                    roomName, activeUsers, lastMessage := chatRoom.GetInfo()
                    fmt.Printf("🔔 %s received update in %s: [Users: %s] Last: %s\n", 
                        u.name, roomName, strings.Join(activeUsers, ", "), lastMessage)
                        
                case <-userCtx.Done():
                    fmt.Printf("👤 %s's listener shutting down\n", u.name)
                    return
                }
            }
        }(user, msgCh, userCancel)
    }

    // Simulate chat activity
    time.Sleep(1 * time.Second)

    // Send some messages
    messages := []struct {
        user *User
        text string
    }{
        {users[0], "Hello everyone!"},
        {users[1], "Hey Alice! How's it going?"},
        {users[2], "Good morning all!"},
        {users[0], "Great to see you all here"},
        {users[1], "This chat system is pretty cool"},
    }

    for i, msg := range messages {
        time.Sleep(2 * time.Second)
        chatRoom.SendMessage(msg.user, msg.text)
        
        // Charlie leaves halfway through
        if i == 2 {
            go func() {
                time.Sleep(1 * time.Second)
                chatRoom.Leave("charlie789")
            }()
        }
    }

    time.Sleep(5 * time.Second)
}
```

## Cache Invalidation

### Distributed Cache Notifications

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"

    "github.com/blackorder/chanhub"
)

type CacheManager struct {
    mu    sync.RWMutex
    data  map[string]interface{}
    hub   *chanhub.Hub
    stats CacheStats
}

type CacheStats struct {
    Hits         int64
    Misses       int64
    Invalidations int64
}

func NewCacheManager() *CacheManager {
    return &CacheManager{
        data: make(map[string]interface{}),
        hub:  chanhub.New(),
    }
}

func (cm *CacheManager) Subscribe(ctx context.Context) <-chan struct{} {
    return cm.hub.Subscribe(ctx)
}

func (cm *CacheManager) Set(key string, value interface{}) {
    cm.mu.Lock()
    cm.data[key] = value
    cm.mu.Unlock()

    fmt.Printf("🔄 Cache SET: %s\n", key)
    cm.hub.Broadcast()
}

func (cm *CacheManager) Get(key string) (interface{}, bool) {
    cm.mu.RLock()
    value, exists := cm.data[key]
    if exists {
        cm.stats.Hits++
    } else {
        cm.stats.Misses++
    }
    cm.mu.RUnlock()

    return value, exists
}

func (cm *CacheManager) Invalidate(key string) {
    cm.mu.Lock()
    if _, exists := cm.data[key]; exists {
        delete(cm.data, key)
        cm.stats.Invalidations++
        cm.mu.Unlock()

        fmt.Printf("❌ Cache INVALIDATE: %s\n", key)
        cm.hub.Broadcast()
    } else {
        cm.mu.Unlock()
    }
}

func (cm *CacheManager) InvalidatePattern(pattern string) {
    cm.mu.Lock()
    var invalidated []string
    for key := range cm.data {
        if strings.Contains(key, pattern) {
            delete(cm.data, key)
            invalidated = append(invalidated, key)
            cm.stats.Invalidations++
        }
    }
    cm.mu.Unlock()

    if len(invalidated) > 0 {
        fmt.Printf("❌ Cache INVALIDATE PATTERN '%s': %v\n", pattern, invalidated)
        cm.hub.Broadcast()
    }
}

func (cm *CacheManager) GetStats() CacheStats {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    return cm.stats
}

// Simulated services that react to cache changes
type Service struct {
    name  string
    cache *CacheManager
}

func (s *Service) Start(ctx context.Context) {
    cacheEvents := s.cache.Subscribe(ctx)
    
    for {
        select {
        case <-cacheEvents:
            stats := s.cache.GetStats()
            fmt.Printf("📊 %s received cache update - Stats: Hits=%d, Misses=%d, Invalidations=%d\n", 
                s.name, stats.Hits, stats.Misses, stats.Invalidations)
                
            // Simulate service-specific reaction to cache changes
            switch s.name {
            case "WebService":
                fmt.Printf("   🌐 WebService: Updating response cache\n")
            case "DatabaseService":
                fmt.Printf("   🗄️  DatabaseService: Checking query cache consistency\n")
            case "MetricsCollector":
                hitRate := float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
                fmt.Printf("   📈 MetricsCollector: Cache hit rate: %.1f%%\n", hitRate)
            }
            
        case <-ctx.Done():
            fmt.Printf("🔌 %s shutting down\n", s.name)
            return
        }
    }
}

func main() {
    cache := NewCacheManager()
    ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
    defer cancel()

    // Start services that monitor cache
    services := []*Service{
        {name: "WebService", cache: cache},
        {name: "DatabaseService", cache: cache},
        {name: "MetricsCollector", cache: cache},
    }

    for _, service := range services {
        go service.Start(ctx)
    }

    time.Sleep(500 * time.Millisecond)

    // Simulate cache operations
    fmt.Println("🚀 Starting cache operations simulation...")

    // Set some initial data
    cache.Set("user:1", map[string]string{"name": "Alice", "email": "alice@example.com"})
    cache.Set("user:2", map[string]string{"name": "Bob", "email": "bob@example.com"})
    cache.Set("config:app", map[string]interface{}{"debug": true, "timeout": 30})
    
    time.Sleep(2 * time.Second)

    // Perform some gets (hits and misses)
    cache.Get("user:1")      // hit
    cache.Get("user:3")      // miss
    cache.Get("config:app")  // hit
    cache.Get("user:2")      // hit
    
    time.Sleep(2 * time.Second)

    // Invalidate specific keys
    cache.Invalidate("user:1")
    
    time.Sleep(2 * time.Second)

    // More operations
    cache.Set("user:3", map[string]string{"name": "Charlie", "email": "charlie@example.com"})
    cache.Get("user:1")      // miss (was invalidated)
    
    time.Sleep(2 * time.Second)

    // Pattern invalidation
    cache.Set("user:4", map[string]string{"name": "David", "email": "david@example.com"})
    cache.Set("session:abc123", map[string]interface{}{"userId": 1, "expires": time.Now().Add(time.Hour)})
    cache.Set("session:def456", map[string]interface{}{"userId": 2, "expires": time.Now().Add(time.Hour)})
    
    time.Sleep(2 * time.Second)
    
    // Invalidate all user data
    cache.InvalidatePattern("user:")
    
    time.Sleep(3 * time.Second)

    // Final stats
    stats := cache.GetStats()
    fmt.Printf("\n📊 Final Cache Statistics:\n")
    fmt.Printf("   Hits: %d\n", stats.Hits)
    fmt.Printf("   Misses: %d\n", stats.Misses)
    fmt.Printf("   Invalidations: %d\n", stats.Invalidations)
    if stats.Hits+stats.Misses > 0 {
        hitRate := float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
        fmt.Printf("   Hit Rate: %.1f%%\n", hitRate)
    }
}
