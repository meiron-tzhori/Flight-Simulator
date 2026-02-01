# Architecture Design Document

## 1. System Architecture

### 1.1 Overview

The Flight Simulator backend follows a **concurrent actor-based architecture** where the simulation engine owns all aircraft state and communicates via channels. This design eliminates shared mutable state and prevents race conditions.

### 1.2 Architecture Diagram

```
                                    ┌─────────────────────┐
                                    │   HTTP Clients      │
                                    └──────────┬──────────┘
                                               │
                                               ▼
┌────────────────────────────────────────────────────────────────────┐
│                          HTTP Server (Gin)                          │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐            │
│  │   /command   │  │    /state    │  │   /stream    │            │
│  │   Handler    │  │   Handler    │  │   Handler    │            │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘            │
│         │                  │                  │                     │
└─────────┼──────────────────┼──────────────────┼────────────────────┘
          │                  │                  │
          │                  │                  │
          ▼                  ▼                  ▼
┌─────────────────┐  ┌────────────────┐  ┌──────────────────┐
│ Command Queue   │  │ State Request  │  │ Subscribe        │
│ (buffered chan) │  │ (request/reply)│  │ (registration)   │
└────────┬────────┘  └────────┬───────┘  └────────┬─────────┘
         │                    │                    │
         │                    │                    │
         └────────────────────┼────────────────────┘
                              │
                              ▼
         ┌──────────────────────────────────────────────────┐
         │         Simulation Engine (Actor)                │
         │                                                   │
         │  ┌────────────────────────────────────────┐      │
         │  │  for { select {                        │      │
         │  │    case <-ticker.C:                    │      │
         │  │      tick() // Update state            │      │
         │  │    case cmd := <-commandQueue:         │      │
         │  │      handleCommand(cmd)                │      │
         │  │    case req := <-stateRequest:         │      │
         │  │      req.reply <- currentState         │      │
         │  │  }}                                     │      │
         │  └────────────────────────────────────────┘      │
         │                                                   │
         │  ┌─────────────────────────────────────┐         │
         │  │  Aircraft State (OWNED)             │         │
         │  │  • Position (lat/lon/alt)           │         │
         │  │  • Velocity (speed, v-speed)        │         │
         │  │  • Heading                           │         │
         │  │  • Active Command                    │         │
         │  └─────────────────────────────────────┘         │
         │                                                   │
         └─────────────────────┬─────────────────────────────┘
                               │
                               │ Publish State
                               ▼
         ┌──────────────────────────────────────────────────┐
         │         State Publisher (PubSub)                 │
         │                                                   │
         │  ┌────────────┐  ┌────────────┐  ┌────────────┐ │
         │  │ Subscriber │  │ Subscriber │  │ Subscriber │ │
         │  │     1      │  │     2      │  │     N      │ │
         │  └──────┬─────┘  └──────┬─────┘  └──────┬─────┘ │
         └─────────┼────────────────┼────────────────┼───────┘
                   │                │                │
                   ▼                ▼                ▼
            SSE Client 1    SSE Client 2    Metrics Collector
```

---

## 2. Component Design

### 2.1 HTTP Server

**Technology**: Gin web framework (or net/http)

**Responsibilities**:
- Request routing
- Request/response serialization
- Middleware (logging, CORS, recovery)
- Graceful shutdown

**Lifecycle**:
```go
type Server struct {
    httpServer *http.Server
    simulator  *Simulator
    logger     *slog.Logger
}

func (s *Server) Start(ctx context.Context) error {
    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        s.httpServer.Shutdown(shutdownCtx)
    }()
    
    return s.httpServer.ListenAndServe()
}
```

### 2.2 Simulation Engine

**Core Principle**: Single-threaded state ownership (Actor model)

**State Ownership**:
```go
type Simulator struct {
    // State (PRIVATE - only accessed in Run goroutine)
    state          AircraftState
    activeCommand  Command
    
    // Communication channels
    commandQueue   chan Command
    stateRequests  chan stateRequest
    publisher      *StatePublisher
    
    // Configuration
    tickInterval   time.Duration
    config         SimConfig
    environment    *Environment
}

type stateRequest struct {
    reply chan AircraftState
}
```

**Main Loop**:
```go
func (s *Simulator) Run(ctx context.Context) error {
    ticker := time.NewTicker(s.tickInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            s.logger.Info("Simulator shutting down")
            return ctx.Err()
        
        case <-ticker.C:
            s.tick()
            s.publisher.Publish(s.state) // Non-blocking
        
        case cmd := <-s.commandQueue:
            if err := s.handleCommand(cmd); err != nil {
                s.logger.Error("Command failed", "error", err)
            }
        
        case req := <-s.stateRequests:
            req.reply <- s.state // Synchronous state query
        }
    }
}
```

**Tick Logic**:
```go
func (s *Simulator) tick() {
    deltaTime := s.tickInterval.Seconds()
    
    // Apply environment effects
    effectiveVelocity := s.state.Velocity
    if s.environment != nil {
        effectiveVelocity = s.environment.ApplyEffects(s.state, effectiveVelocity)
    }
    
    // Execute active command
    switch cmd := s.activeCommand.(type) {
    case *GoToCommand:
        s.executeGoTo(cmd, deltaTime, effectiveVelocity)
    case *TrajectoryCommand:
        s.executeTrajectory(cmd, deltaTime, effectiveVelocity)
    case *HoldCommand:
        s.executeHold(deltaTime)
    }
    
    s.state.Timestamp = time.Now()
}
```

### 2.3 Command Queue

**Design**: Buffered channel for backpressure handling

```go
const defaultCommandQueueSize = 100

commandQueue := make(chan Command, defaultCommandQueueSize)
```

**Submission Pattern** (non-blocking with timeout):
```go
func (s *Simulator) SubmitCommand(ctx context.Context, cmd Command) error {
    select {
    case s.commandQueue <- cmd:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(5 * time.Second):
        return ErrCommandQueueFull
    }
}
```

**Metrics**:
- Queue depth gauge
- Commands accepted/rejected counters

### 2.4 State Publisher (PubSub)

**Design**: Fan-out with buffered channels per subscriber

```go
type StatePublisher struct {
    mu          sync.RWMutex
    subscribers map[string]chan AircraftState
    bufferSize  int
}

func (p *StatePublisher) Subscribe(id string) <-chan AircraftState {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    ch := make(chan AircraftState, p.bufferSize)
    p.subscribers[id] = ch
    return ch
}

func (p *StatePublisher) Unsubscribe(id string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if ch, exists := p.subscribers[id]; exists {
        close(ch)
        delete(p.subscribers, id)
    }
}

func (p *StatePublisher) Publish(state AircraftState) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    for id, ch := range p.subscribers {
        select {
        case ch <- state:
            // Sent successfully
        default:
            // Channel full, skip (or log warning)
            log.Warn("Subscriber buffer full", "id", id)
        }
    }
}
```

**Subscriber Cleanup**:
```go
func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    subID := uuid.New().String()
    stateChan := h.publisher.Subscribe(subID)
    defer h.publisher.Unsubscribe(subID)
    
    // SSE setup...
    
    for {
        select {
        case state := <-stateChan:
            // Send SSE event
        case <-r.Context().Done():
            return
        }
    }
}
```

### 2.5 Environment Module

**Design**: Modular effects that can be enabled/disabled

```go
type Environment struct {
    wind     *WindEffect
    humidity *HumidityEffect
    terrain  *TerrainMap
}

type Effect interface {
    Apply(state AircraftState, velocity Velocity) Velocity
}

func (e *Environment) ApplyEffects(state AircraftState, velocity Velocity) Velocity {
    result := velocity
    
    if e.wind != nil {
        result = e.wind.Apply(state, result)
    }
    
    if e.humidity != nil {
        result = e.humidity.Apply(state, result)
    }
    
    return result
}
```

**Wind Effect**:
```go
type WindEffect struct {
    direction float64 // degrees
    speed     float64 // m/s
}

func (w *WindEffect) Apply(state AircraftState, velocity Velocity) Velocity {
    // Convert wind to velocity components
    windNorth := w.speed * math.Cos(toRadians(w.direction))
    windEast := w.speed * math.Sin(toRadians(w.direction))
    
    // Convert aircraft heading to velocity components
    acNorth := velocity.GroundSpeed * math.Cos(toRadians(state.Heading))
    acEast := velocity.GroundSpeed * math.Sin(toRadians(state.Heading))
    
    // Add wind effect
    totalNorth := acNorth + windNorth
    totalEast := acEast + windEast
    
    // Calculate new ground speed and track
    newSpeed := math.Sqrt(totalNorth*totalNorth + totalEast*totalEast)
    
    return Velocity{
        GroundSpeed:   newSpeed,
        VerticalSpeed: velocity.VerticalSpeed,
    }
}
```

---

## 3. Concurrency Patterns

### 3.1 Actor Model Implementation

**Key Principles**:
1. **Single owner**: Only the simulation goroutine mutates state
2. **Message passing**: All communication via channels
3. **Immutable messages**: State copies for reading
4. **No shared state**: Each component owns its data

**Benefits**:
- No mutexes needed
- No race conditions possible
- Simple to reason about
- Easy to test

### 3.2 Request-Reply Pattern

**State Query**:
```go
// Handler
func (h *StateHandler) GetState() (AircraftState, error) {
    req := stateRequest{
        reply: make(chan AircraftState, 1),
    }
    
    select {
    case h.simulator.stateRequests <- req:
        state := <-req.reply
        return state, nil
    case <-time.After(1 * time.Second):
        return AircraftState{}, ErrTimeout
    }
}

// Simulator
case req := <-s.stateRequests:
    req.reply <- s.state // Send immutable copy
```

### 3.3 Context-Based Cancellation

**Shutdown Sequence**:
```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    var wg sync.WaitGroup
    
    // Start simulator
    wg.Add(1)
    go func() {
        defer wg.Done()
        simulator.Run(ctx)
    }()
    
    // Start HTTP server
    wg.Add(1)
    go func() {
        defer wg.Done()
        server.Start(ctx)
    }()
    
    // Wait for signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
    
    // Initiate shutdown
    cancel()
    
    // Wait for graceful shutdown
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        log.Info("Graceful shutdown complete")
    case <-time.After(30 * time.Second):
        log.Error("Shutdown timeout exceeded")
    }
}
```

### 3.4 Goroutine Leak Prevention

**WaitGroup Tracking**:
```go
type Simulator struct {
    wg sync.WaitGroup
}

func (s *Simulator) startWorker(ctx context.Context, name string, fn func(context.Context)) {
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        defer func() {
            if r := recover(); r != nil {
                s.logger.Error("Worker panic", "worker", name, "panic", r)
            }
        }()
        fn(ctx)
    }()
}

func (s *Simulator) Shutdown() error {
    done := make(chan struct{})
    go func() {
        s.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-time.After(10 * time.Second):
        return ErrShutdownTimeout
    }
}
```

---

## 4. Data Flow

### 4.1 Command Flow

```
HTTP POST /command/goto
  |
  ├─> Request Validation (API layer)
  |
  ├─> Command Creation (models)
  |
  ├─> Submit to Queue (non-blocking with timeout)
  |     |
  |     └─> commandQueue channel
  |           |
  |           └─> Simulator.Run() select case
  |                 |
  |                 ├─> Validate Command
  |                 |
  |                 ├─> Store as activeCommand
  |                 |
  |                 └─> Log acceptance
  |
  └─> Return 200 OK to client

Next Tick:
  |
  └─> Simulator.tick()
        |
        ├─> Read activeCommand
        |
        ├─> Calculate new position
        |
        ├─> Update state
        |
        └─> Publish state
              |
              └─> StatePublisher.Publish()
                    |
                    └─> Fan out to all subscribers
```

### 4.2 State Query Flow

```
HTTP GET /state
  |
  ├─> Create stateRequest with reply channel
  |
  ├─> Send request to simulator
  |     |
  |     └─> stateRequests channel
  |           |
  |           └─> Simulator.Run() select case
  |                 |
  |                 └─> Send current state to reply channel
  |
  ├─> Wait for reply (with timeout)
  |
  ├─> Receive state
  |
  └─> Return JSON to client
```

### 4.3 SSE Stream Flow

```
HTTP GET /stream
  |
  ├─> SSE headers setup
  |
  ├─> Subscribe to state publisher
  |     |
  |     └─> StatePublisher.Subscribe(uuid)
  |           |
  |           └─> Create buffered channel
  |                 |
  |                 └─> Add to subscribers map
  |
  └─> Stream loop:
        |
        ├─> for state := range stateChan
        |     |
        |     ├─> Format SSE event
        |     |
        |     └─> Write to response
        |
        └─> On client disconnect:
              |
              └─> Unsubscribe (cleanup)
```

---

## 5. Package Structure

```
internal/
├── api/
│   ├── server.go           # HTTP server setup
│   ├── handlers/
│   │   ├── command.go      # Command endpoints
│   │   ├── state.go        # State query endpoint
│   │   ├── stream.go       # SSE streaming
│   │   └── health.go       # Health check
│   ├── middleware/
│   │   ├── logging.go      # Request logging
│   │   ├── recovery.go     # Panic recovery
│   │   └── cors.go         # CORS headers
│   └── validation/
│       └── validate.go     # Request validation
│
├── simulator/
│   ├── simulator.go        # Main simulation engine
│   ├── physics.go          # Movement calculations
│   ├── commands.go         # Command execution logic
│   └── config.go           # Simulator configuration
│
├── pubsub/
│   └── publisher.go        # State publisher implementation
│
├── models/
│   ├── aircraft.go         # Aircraft state
│   ├── command.go          # Command types
│   └── errors.go           # Error types
│
├── environment/
│   ├── environment.go      # Environment coordinator
│   ├── wind.go             # Wind effect
│   ├── humidity.go         # Humidity effect
│   └── terrain.go          # Terrain map (bonus)
│
├── config/
│   ├── config.go           # Configuration structs
│   └── loader.go           # Config file loading
│
└── observability/
    ├── logger.go           # Structured logging
    └── metrics.go          # Prometheus metrics

pkg/
└── geo/
    ├── distance.go         # Haversine distance
    ├── bearing.go          # Bearing calculations
    └── conversions.go      # Coordinate conversions
```

---

## 6. Thread Safety Analysis

### 6.1 Thread-Safe Components

**StatePublisher** (Mutex-protected)
```go
type StatePublisher struct {
    mu          sync.RWMutex  // Protects subscribers map
    subscribers map[string]chan AircraftState
}

// RWMutex allows multiple concurrent reads, exclusive writes
// Safe for concurrent Subscribe/Unsubscribe/Publish
```

**Config** (Immutable after load)
```go
type Config struct {
    // Read-only after initialization
    // No synchronization needed
}
```

### 6.2 Race-Free Components

**Simulator State** (Single owner)
```go
// SAFE: Only accessed in Run() goroutine
type Simulator struct {
    state AircraftState  // No synchronization needed!
}
```

**Channels** (Built-in synchronization)
```go
// SAFE: Channels provide synchronization
commandQueue   chan Command
stateRequests  chan stateRequest
```

---

## 7. Error Handling Strategy

### 7.1 Error Types

```go
var (
    ErrInvalidCoordinates  = errors.New("invalid coordinates")
    ErrInvalidSpeed        = errors.New("invalid speed")
    ErrCommandQueueFull    = errors.New("command queue full")
    ErrSimulatorNotRunning = errors.New("simulator not running")
    ErrTerrainConflict     = errors.New("terrain collision detected")
    ErrTimeout             = errors.New("operation timeout")
)
```

### 7.2 Error Propagation

```go
// API Layer: Translate to HTTP codes
func (h *Handler) HandleCommand(c *gin.Context) {
    err := h.simulator.SubmitCommand(c.Request.Context(), cmd)
    
    switch {
    case errors.Is(err, ErrInvalidCoordinates):
        c.JSON(400, ErrorResponse{Error: err.Error()})
    case errors.Is(err, ErrCommandQueueFull):
        c.JSON(503, ErrorResponse{Error: "Service busy"})
    case err != nil:
        c.JSON(500, ErrorResponse{Error: "Internal error"})
    default:
        c.JSON(200, SuccessResponse{})
    }
}
```

### 7.3 Panic Recovery

```go
func RecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                log.Error("Panic recovered", "panic", r, "stack", debug.Stack())
                c.JSON(500, ErrorResponse{Error: "Internal server error"})
            }
        }()
        c.Next()
    }
}
```

---

## 8. Performance Considerations

### 8.1 Tick Rate vs CPU Usage

**Trade-off**:
- Higher tick rate → smoother simulation → more CPU
- Lower tick rate → choppy movement → less CPU

**Recommendation**: 30Hz (33ms ticks)
- Good balance for demo purposes
- ~3% CPU usage on modern hardware
- Smooth enough for visualization

### 8.2 Channel Buffer Sizing

**Command Queue**: 100 commands
- Handles burst traffic
- Provides backpressure signal
- Minimal memory overhead (~10KB)

**State Publisher**: 10 states per subscriber
- Prevents blocking on slow consumers
- Allows brief network hiccups
- Auto-drops old states if subscriber lags

### 8.3 Memory Usage

**Baseline**: ~10MB
- Golang runtime: ~5MB
- Application code: ~2MB
- Buffers and channels: ~1MB
- Per-connection overhead: ~50KB

**With 100 SSE clients**: ~15MB
- 100 × 50KB = 5MB additional

---

## 9. Testing Strategy

### 9.1 Unit Tests

```go
// Test simulation physics
func TestSimulatorGoTo(t *testing.T) {
    sim := NewSimulator(defaultConfig)
    
    cmd := &GoToCommand{
        Target: Position{Lat: 32.1, Lon: 34.8, Alt: 1000},
        Speed:  100,
    }
    
    // Run simulation for 10 seconds
    for i := 0; i < 300; i++ { // 30Hz × 10s
        sim.tick()
    }
    
    // Assert aircraft reached target
    assert.Near(t, sim.state.Position.Lat, 32.1, 0.001)
}
```

### 9.2 Concurrency Tests

```bash
# Race detector
go test -race -count=100 ./...

# Stress test
go test -parallel=20 -count=1000 ./internal/simulator
```

### 9.3 Integration Tests

```go
func TestEndToEnd(t *testing.T) {
    // Start server
    server := setupTestServer(t)
    defer server.Shutdown()
    
    // Send goto command
    resp := httpPost("/command/goto", gotoPayload)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Wait for movement
    time.Sleep(100 * time.Millisecond)
    
    // Check state
    state := httpGet("/state")
    assert.NotEqual(t, initialPosition, state.Position)
}
```

---

## 10. Deployment Considerations

### 10.1 Build

```bash
# Static binary
CGO_ENABLED=0 go build -o bin/simulator cmd/simulator/main.go

# With version info
go build -ldflags "-X main.version=$(git describe --tags)" -o bin/simulator
```

### 10.2 Runtime Requirements

- **CPU**: 1 core, 5% average usage
- **Memory**: 20MB baseline, 50KB per SSE client
- **Network**: Minimal (HTTP REST only)
- **Disk**: None (stateless service)

### 10.3 Configuration

```yaml
# Environment variable override
export SIM_TICK_RATE_HZ=60
export SIM_PORT=8080
export SIM_LOG_LEVEL=debug

./bin/simulator
```

---

## 11. Future Enhancements

### 11.1 Persistence
- Save/restore simulation state
- Command history logging
- Replay capability

### 11.2 Advanced Features
- Multiple aircraft support
- Collision avoidance
- Realistic flight dynamics
- Weather API integration

### 11.3 Scalability
- Horizontal scaling (stateless design)
- Redis-based state broadcasting
- Load balancer support
- Kubernetes deployment

---

## 12. References

### 12.1 Go Concurrency Patterns
- [Go Concurrency Patterns (2012)](https://www.youtube.com/watch?v=f6kdp27TYZs)
- [Advanced Go Concurrency Patterns](https://www.youtube.com/watch?v=QDDwwePbDtw)
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)

### 12.2 Actor Model
- [The Actor Model in 10 Minutes](https://www.brianstorti.com/the-actor-model/)
- [Go's approach to actors](https://changelog.com/gotime/11)

### 12.3 Geodetic Calculations
- [Haversine formula](https://en.wikipedia.org/wiki/Haversine_formula)
- [Aviation formulary](http://www.edwilliams.org/avform147.htm)
