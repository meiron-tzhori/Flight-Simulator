# Product Requirements Document: Airborne Flight Simulator Backend

## 1. Overview

### 1.1 Project Goal
Build a concurrent backend service in Go that simulates aircraft flight and provides command & control capabilities through a REST API. The system emphasizes concurrency patterns, architectural clarity, and robustness over aerodynamic accuracy.

### 1.2 Success Criteria
- **Correctness**: Bug-free simulation logic with proper state management
- **Concurrency**: Senior-level concurrent architecture using Go patterns
- **Architecture**: Clean separation of concerns with modular design
- **Robustness**: Comprehensive error handling and observability
- **Clarity**: Well-documented, readable code with clear design decisions

### 1.3 Non-Goals
- Realistic aerodynamic modeling
- Complex physics simulation
- Production-ready deployment infrastructure
- Authentication/authorization mechanisms

---

## 2. System Architecture

### 2.1 High-Level Components

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP REST API                          │
│  (/command/goto, /command/trajectory, /state, /stream)      │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   Command Handler                            │
│         (Validates & routes commands to queue)               │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   Command Queue                              │
│              (Buffered channel)                              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Simulation Engine (Actor)                       │
│  • Tick loop (goroutine)                                     │
│  • State ownership                                           │
│  • Physics calculations                                      │
│  • Command execution                                         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              State Publisher (PubSub)                        │
│  • Broadcast state updates                                   │
│  • Multiple subscriber support                               │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ├──────► SSE Stream Handler
                     │
                     └──────► State Query Handler
```

### 2.2 Component Responsibilities

#### HTTP REST API Layer
- Accept incoming HTTP requests
- Request validation and parsing
- Response formatting
- SSE connection management (bonus)

#### Command Handler
- Command validation (coordinates, speed limits)
- Command transformation to internal format
- Synchronous command submission to queue
- Error propagation to API layer

#### Command Queue
- Buffered channel for command ingestion
- Decouples API from simulation engine
- Provides backpressure handling
- FIFO command processing

#### Simulation Engine (Actor)
- **Single owner** of aircraft state (race-free)
- Continuous tick loop at configurable Hz
- Command processing from queue
- Physics/movement calculations
- State publication after each tick
- Environment effects integration (bonus)

#### State Publisher (PubSub)
- Fan-out state updates to multiple subscribers
- Non-blocking broadcast
- Subscriber registration/cleanup
- Buffer overflow handling

#### Environment Module (Bonus)
- Wind vector calculations
- Humidity effects
- Terrain collision detection
- Modular on/off configuration

---

## 3. Data Models

### 3.1 Aircraft State

```go
type AircraftState struct {
    Position  Position  `json:"position"`
    Velocity  Velocity  `json:"velocity"`
    Heading   float64   `json:"heading"`    // degrees, 0-360
    Timestamp time.Time `json:"timestamp"`
}

type Position struct {
    Latitude  float64 `json:"latitude"`   // degrees, -90 to 90
    Longitude float64 `json:"longitude"`  // degrees, -180 to 180
    Altitude  float64 `json:"altitude"`   // meters MSL
}

type Velocity struct {
    GroundSpeed float64 `json:"ground_speed"` // m/s
    VerticalSpeed float64 `json:"vertical_speed"` // m/s
}
```

### 3.2 Commands

```go
type CommandType string

const (
    CommandTypeGoTo       CommandType = "goto"
    CommandTypeTrajectory CommandType = "trajectory"
    CommandTypeStop       CommandType = "stop"  // bonus
    CommandTypeHold       CommandType = "hold"  // bonus
)

type Command struct {
    Type      CommandType `json:"type"`
    GoTo      *GoToCommand      `json:"goto,omitempty"`
    Trajectory *TrajectoryCommand `json:"trajectory,omitempty"`
}

type GoToCommand struct {
    Target Position `json:"target"`
    Speed  *float64 `json:"speed,omitempty"` // m/s, optional
}

type TrajectoryCommand struct {
    Waypoints []Waypoint `json:"waypoints"`
    Loop      bool       `json:"loop"`
}

type Waypoint struct {
    Position Position `json:"position"`
    Speed    *float64 `json:"speed,omitempty"` // m/s, optional
}
```

### 3.3 Environment State (Bonus)

```go
type Environment struct {
    Wind     *WindVector `json:"wind,omitempty"`
    Humidity *float64    `json:"humidity,omitempty"` // 0-100%
    Terrain  *TerrainMap `json:"-"`
}

type WindVector struct {
    Direction float64 `json:"direction"` // degrees
    Speed     float64 `json:"speed"`     // m/s
}
```

---

## 4. API Specification

### 4.1 Core Endpoints

#### POST /command/goto
Submit a "go to point" command.

**Request:**
```json
{
  "lat": 32.0853,
  "lon": 34.7818,
  "alt": 1000.0,
  "speed": 100.0
}
```

**Response (200 OK):**
```json
{
  "status": "accepted",
  "command_id": "cmd-123",
  "message": "Go-to command accepted"
}
```

**Errors:**
- 400: Invalid coordinates or parameters
- 503: Command queue full

#### POST /command/trajectory
Submit a trajectory with multiple waypoints.

**Request:**
```json
{
  "waypoints": [
    {"lat": 32.0853, "lon": 34.7818, "alt": 1000.0, "speed": 100.0},
    {"lat": 32.1000, "lon": 34.8000, "alt": 1500.0},
    {"lat": 32.1200, "lon": 34.8200, "alt": 2000.0}
  ],
  "loop": false
}
```

**Response (200 OK):**
```json
{
  "status": "accepted",
  "command_id": "cmd-124",
  "waypoint_count": 3,
  "message": "Trajectory command accepted"
}
```

#### GET /state
Query current aircraft state.

**Response (200 OK):**
```json
{
  "position": {
    "latitude": 32.0853,
    "longitude": 34.7818,
    "altitude": 1000.0
  },
  "velocity": {
    "ground_speed": 100.0,
    "vertical_speed": 5.0
  },
  "heading": 45.0,
  "timestamp": "2026-02-01T18:39:00Z",
  "active_command": "goto",
  "eta_seconds": 120.5
}
```

#### GET /health
Health check endpoint.

**Response (200 OK):**
```json
{
  "status": "healthy",
  "simulation_running": true,
  "tick_rate_hz": 30.0,
  "uptime_seconds": 3600
}
```

### 4.2 Bonus Endpoints

#### GET /stream
Server-Sent Events (SSE) stream of aircraft state updates.

**Response Headers:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Event Stream:**
```
data: {"position":{...},"velocity":{...},"heading":45.0,"timestamp":"..."}

data: {"position":{...},"velocity":{...},"heading":46.0,"timestamp":"..."}

...
```

#### POST /command/stop
Stop the aircraft immediately (hold current position).

**Response (200 OK):**
```json
{
  "status": "accepted",
  "message": "Aircraft stopped"
}
```

#### POST /command/hold
Hold current altitude and position (orbit).

**Response (200 OK):**
```json
{
  "status": "accepted",
  "message": "Aircraft holding position"
}
```

---

## 5. Concurrency Architecture

### 5.1 Concurrency Principles

**Actor Model for State Ownership:**
- Simulation engine is the **single owner** of aircraft state
- All state mutations happen in simulation goroutine
- No shared mutable state → no race conditions
- Communication via channels only

**Channel-Based Communication:**
- Command ingestion: buffered channel (capacity: 100)
- State broadcasting: fan-out via pubsub
- Graceful shutdown: context cancellation

**Lifecycle Management:**
- Context-based cancellation throughout
- WaitGroup for goroutine tracking
- Clean shutdown sequence with timeouts

### 5.2 Goroutine Structure

```go
// Main goroutines in the system:

1. HTTP Server Goroutine
   - Managed by http.Server
   - Handles incoming requests
   - Context-aware shutdown

2. Simulation Engine Goroutine
   - Ticker-based loop (30Hz default)
   - Owns aircraft state
   - Processes commands from queue
   - Publishes state updates

3. State Publisher Goroutines (1 per subscriber)
   - Receives state from publisher
   - Forwards to SSE client or query cache
   - Cleanup on disconnect

4. Metrics Collection Goroutine (optional)
   - Periodic metrics aggregation
   - Prometheus exporter
```

### 5.3 Race Condition Prevention

**State Access Pattern:**
```
Read State:  API Handler → Request via channel → Simulation responds
Write State: Only simulation goroutine mutates state
Broadcast:   Simulation → PubSub → Subscribers (read-only copies)
```

**No Mutexes Needed:**
- Actor model ensures single-threaded state access
- Channels provide synchronization
- Immutable state copies for broadcasting

---

## 6. Simulation Logic

### 6.1 Tick Loop

```go
func (s *Simulator) Run(ctx context.Context) error {
    ticker := time.NewTicker(s.tickInterval) // e.g., 33ms for 30Hz
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        
        case <-ticker.C:
            s.tick()
        
        case cmd := <-s.commandQueue:
            s.handleCommand(cmd)
        }
    }
}
```

### 6.2 Movement Calculation (Simplified)

**Go-To-Point Logic:**
1. Calculate bearing from current position to target
2. Calculate distance to target
3. If distance < tolerance (e.g., 50m): command complete
4. Otherwise:
   - Update heading toward target
   - Move at current speed
   - Adjust altitude gradually

**Trajectory Logic:**
1. Identify current waypoint
2. Apply go-to-point logic for current waypoint
3. When waypoint reached, advance to next
4. If loop enabled and last waypoint reached, restart

**Physics Calculations:**
```go
// Simplified 2D movement
distance := speed * deltaTime
newLat := currentLat + (distance * cos(heading)) / metersPerDegreeLat
newLon := currentLon + (distance * sin(heading)) / metersPerDegreeLon

// Altitude change
altitudeDelta := verticalSpeed * deltaTime
newAlt := currentAlt + altitudeDelta
```

### 6.3 Environment Effects (Bonus)

**Wind Integration:**
```go
// Add wind vector to ground track
groundTrack := heading + windEffect
effectiveSpeed := speed + windSpeedComponent
```

**Humidity Effects:**
```go
// Reduce max climb rate in high humidity
if humidity > 80 {
    maxVerticalSpeed *= 0.8
}
```

**Terrain Collision Avoidance:**
```go
// Check terrain altitude at current position
terrainAlt := terrain.GetAltitude(lat, lon)
minSafeAlt := terrainAlt + safetyMargin // e.g., 100m

if targetAlt < minSafeAlt {
    return ErrTerrainConflict
}
```

---

## 7. Configuration

### 7.1 Configuration Parameters

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 5s
  write_timeout: 10s
  shutdown_timeout: 30s

simulation:
  tick_rate_hz: 30
  command_queue_size: 100
  
  # Initial state
  initial_position:
    latitude: 32.0853
    longitude: 34.7818
    altitude: 1000.0
  
  initial_velocity:
    ground_speed: 50.0
    vertical_speed: 0.0
  
  initial_heading: 0.0
  
  # Flight parameters
  default_speed: 100.0  # m/s
  max_speed: 250.0
  max_climb_rate: 15.0  # m/s
  max_descent_rate: 10.0
  
  # Tolerances
  position_tolerance: 50.0  # meters
  heading_change_rate: 5.0   # degrees per second

environment:
  enabled: true
  
  wind:
    enabled: true
    direction: 270  # degrees
    speed: 10.0     # m/s
  
  humidity:
    enabled: false
    value: 60.0  # percent
  
  terrain:
    enabled: false
    safety_margin: 100.0  # meters

logging:
  level: "info"  # debug, info, warn, error
  format: "json"

metrics:
  enabled: true
  port: 9090
```

---

## 8. Error Handling

### 8.1 Error Categories

**Validation Errors (400 Bad Request):**
- Invalid coordinates (lat/lon out of range)
- Invalid altitude (negative)
- Invalid speed (negative or exceeds max)
- Empty waypoint list
- Malformed JSON

**Service Errors (503 Service Unavailable):**
- Command queue full
- Simulation not running
- Shutdown in progress

**Terrain Errors (422 Unprocessable Entity):**
- Command would cause terrain collision
- Insufficient altitude for safety margin

### 8.2 Error Response Format

```json
{
  "error": {
    "code": "INVALID_COORDINATES",
    "message": "Latitude must be between -90 and 90 degrees",
    "field": "lat",
    "value": 120.5
  }
}
```

---

## 9. Observability

### 9.1 Logging

**Structured Logging Fields:**
- `timestamp`: ISO8601 timestamp
- `level`: debug, info, warn, error
- `component`: api, simulator, pubsub, etc.
- `event`: command_received, tick_complete, state_published
- `context`: relevant contextual data

**Log Events:**
- Server start/stop
- Command received/accepted/rejected
- Simulation tick (debug level)
- State published (debug level)
- Errors and warnings
- Client connections/disconnections

### 9.2 Metrics (Bonus)

**Prometheus Metrics:**
```
# Counters
commands_total{type="goto|trajectory|stop|hold", status="accepted|rejected"}
simulation_ticks_total
state_publishes_total

# Gauges
active_sse_clients
command_queue_depth
aircraft_altitude_meters
aircraft_speed_mps

# Histograms
command_processing_duration_seconds
tick_duration_seconds
api_request_duration_seconds
```

---

## 10. Testing Strategy

### 10.1 Unit Tests

- Position/velocity calculations
- Bearing calculations
- Distance calculations
- Command validation
- Environment effect calculations

### 10.2 Integration Tests

- End-to-end API flows
- Command submission → state change
- Trajectory following
- SSE streaming
- Graceful shutdown

### 10.3 Concurrency Tests

```bash
# Run with race detector
go test -race ./...

# Stress test with parallel requests
go test -parallel 10 -count 100
```

### 10.4 Performance Tests

- Tick rate stability under load
- Command queue saturation
- SSE client scalability (100+ concurrent)
- Memory leak detection

---

## 11. Deliverables

### 11.1 Repository Structure

```
flight-simulator/
├── README.md                 # Build, run, usage instructions
├── docs/
│   ├── PRD.md               # This document
│   ├── ARCHITECTURE.md      # Detailed architecture
│   ├── API.md               # API documentation
│   └── DESIGN_DECISIONS.md  # Assumptions & tradeoffs
├── cmd/
│   └── simulator/
│       └── main.go          # Entry point
├── internal/
│   ├── api/                 # HTTP handlers
│   ├── simulator/           # Simulation engine
│   ├── pubsub/              # State publisher
│   ├── models/              # Data models
│   ├── environment/         # Environment effects
│   └── config/              # Configuration
├── pkg/
│   └── geo/                 # Geographic utilities
├── configs/
│   └── config.yaml          # Default configuration
├── scripts/
│   ├── demo.sh              # Demo script
│   └── curl-examples.sh     # API examples
├── tests/
│   ├── integration/
│   └── load/
├── go.mod
└── go.sum
```

### 11.2 Documentation Requirements

**README.md must include:**
1. Quick start (build & run)
2. API usage with curl examples
3. Configuration options
4. Architecture overview (brief)
5. Testing instructions
6. Key design decisions

**DESIGN_DECISIONS.md must include:**
1. Assumptions made
2. Tradeoffs considered
3. Alternative approaches rejected
4. Future improvements
5. Known limitations

---

## 12. Timeline & Milestones

### Phase 1: Core Foundation (Day 1-2)
- [ ] Project setup and structure
- [ ] Data models implementation
- [ ] Configuration system
- [ ] Basic simulation engine (single goroutine)
- [ ] Movement calculations

### Phase 2: Concurrency (Day 2-3)
- [ ] Actor-based simulation engine
- [ ] Command queue implementation
- [ ] State publisher (pubsub)
- [ ] Context-based lifecycle
- [ ] Race condition testing

### Phase 3: API (Day 3-4)
- [ ] HTTP server setup
- [ ] Core endpoints (/command/goto, /trajectory, /state)
- [ ] Health endpoint
- [ ] Error handling
- [ ] Request validation

### Phase 4: Testing & Refinement (Day 4-5)
- [ ] Unit tests
- [ ] Integration tests
- [ ] Race detector validation
- [ ] Performance testing
- [ ] Documentation

### Phase 5: Bonus Features (Day 5-6)
- [ ] SSE streaming endpoint
- [ ] Stop/Hold commands
- [ ] Wind effects
- [ ] Metrics/observability
- [ ] Demo script

---

## 13. Acceptance Criteria

### Must Have
- ✅ Aircraft state tracks position, velocity, heading, timestamp
- ✅ Simulation runs continuously at configurable tick rate
- ✅ REST API accepts goto and trajectory commands
- ✅ GET /state returns current aircraft state
- ✅ **Proper concurrency**: actor model, channels, contexts
- ✅ Race-free state management
- ✅ Clean shutdown with graceful termination
- ✅ Comprehensive README with examples
- ✅ No race conditions (verified with `-race`)

### Should Have
- ⭐ SSE streaming endpoint
- ⭐ Stop/Hold commands
- ⭐ At least one environment effect (wind recommended)
- ⭐ Structured logging
- ⭐ Prometheus metrics
- ⭐ Integration tests

### Nice to Have
- 💎 Multiple environment effects
- 💎 Terrain collision avoidance
- 💎 Load testing results
- 💎 Performance benchmarks
- 💎 Docker support
- 💎 Makefile for common tasks

---

## 14. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Race conditions in state access | High | Actor model, single state owner, race detector |
| Command queue saturation | Medium | Buffered channel, backpressure, queue depth metrics |
| Goroutine leaks | Medium | Context cancellation, WaitGroup tracking, testing |
| SSE client memory leaks | Medium | Proper cleanup, connection tracking, timeouts |
| Simulation drift at high tick rates | Low | Use time.Ticker, measure actual delta time |
| Complex trajectory edge cases | Low | Comprehensive unit tests, clear tolerance definitions |

---

## 15. Open Questions

1. **Coordinate system**: Should we use geodetic (lat/lon) or Cartesian (x/y/z)?
   - **Decision**: Geodetic for realism, simplified calculations acceptable

2. **Command queueing**: FIFO or priority-based?
   - **Decision**: FIFO (simpler), priority as future enhancement

3. **State update frequency for SSE**: Same as tick rate or throttled?
   - **Decision**: Throttled to 10Hz to reduce bandwidth

4. **Waypoint transition**: Sharp turn or smooth interpolation?
   - **Decision**: Simple turn-to-heading, smoothing as enhancement

5. **Error recovery**: What happens on invalid commands mid-flight?
   - **Decision**: Reject command, continue current command

---

## 16. Success Metrics

The assignment will be evaluated on:

1. **Correctness** (30%)
   - Simulation logic works as specified
   - Commands produce expected behavior
   - No bugs in core functionality

2. **Concurrency & Architecture** (40%)
   - Proper use of goroutines and channels
   - Actor model implementation
   - Race-free design (verified)
   - Clean separation of concerns

3. **Robustness** (15%)
   - Error handling
   - Graceful shutdown
   - Edge case handling
   - No panics or crashes

4. **Observability** (10%)
   - Logging quality
   - Metrics (if implemented)
   - Debuggability

5. **Code Quality** (5%)
   - Readability
   - Documentation
   - Project organization
   - Testing coverage
