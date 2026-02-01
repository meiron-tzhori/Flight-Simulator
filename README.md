# Flight Simulator Backend

> A concurrent backend service simulating aircraft flight with command & control capabilities.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Overview

This is a home assignment implementation for a Senior Developer position, showcasing:

- **Concurrent architecture** using Go's actor model pattern
- **REST API** for aircraft command & control
- **Real-time state streaming** via Server-Sent Events (SSE)
- **Race-free design** verified with Go's race detector
- **Production-like patterns** including graceful shutdown and observability

### Key Features

- ✈️ Simulated aircraft with position, velocity, and heading
- 🎯 Command-based control (go-to-point, trajectories)
- 📡 Real-time state streaming
- 🌪️ Environmental effects (wind, terrain - bonus)
- 🔄 Concurrent design with channels and actors
- 🛡️ Comprehensive error handling
- 📊 Health checks and metrics (bonus)

## Quick Start

### Prerequisites

- Go 1.22 or higher
- `curl` and `jq` (for testing)

### Installation

```bash
# Clone the repository
git clone https://github.com/meiron-tzhori/Flight-Simulator.git
cd Flight-Simulator

# Download dependencies
go mod download

# Build the binary
go build -o bin/simulator cmd/simulator/main.go
```

### Running the Simulator

```bash
# Run with default configuration
go run cmd/simulator/main.go

# Or use the built binary
./bin/simulator

# Run with custom config
go run cmd/simulator/main.go -config configs/config.yaml

# Override via environment variables
export SIM_TICK_RATE_HZ=60
export SIM_PORT=8080
go run cmd/simulator/main.go
```

The server will start on `http://localhost:8080`.

## API Usage

### Health Check

```bash
curl http://localhost:8080/health
```

### Get Aircraft State

```bash
curl http://localhost:8080/state | jq .
```

### Send Go-To Command

```bash
curl -X POST http://localhost:8080/command/goto \
  -H "Content-Type: application/json" \
  -d '{
    "lat": 32.0853,
    "lon": 34.7818,
    "alt": 1000.0,
    "speed": 100.0
  }'
```

### Send Trajectory Command

```bash
curl -X POST http://localhost:8080/command/trajectory \
  -H "Content-Type: application/json" \
  -d '{
    "waypoints": [
      {"lat": 32.0853, "lon": 34.7818, "alt": 1000.0},
      {"lat": 32.1053, "lon": 34.7818, "alt": 1500.0},
      {"lat": 32.0953, "lon": 34.8018, "alt": 1500.0}
    ]
  }'
```

### Stream State Updates (SSE)

```bash
# Stream to console
curl -N http://localhost:8080/stream

# Stream with formatted output
curl -N http://localhost:8080/stream | \
  grep '^data:' | \
  sed 's/^data: //' | \
  jq -c '{lat: .position.latitude, lon: .position.longitude, alt: .position.altitude}'
```

### Interactive Examples

```bash
# Run interactive demo script
chmod +x scripts/curl-examples.sh
./scripts/curl-examples.sh

# Or run all examples
./scripts/curl-examples.sh --all
```

## Testing

### Run All Tests

```bash
# Unit tests
go test ./...

# With race detector (IMPORTANT!)
go test -race ./...

# With coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run Specific Test Suites

```bash
# Test simulation engine
go test -v ./internal/simulator/...

# Test API handlers
go test -v ./internal/api/...

# Test with race detector and multiple iterations
go test -race -count=100 ./internal/simulator
```

### Integration Tests

```bash
# Run integration tests
go test -v -tags=integration ./tests/integration/...
```

## Configuration

Configuration can be provided via YAML file or environment variables.

### Default Configuration

See `configs/config.yaml` for all available options.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|----------|
| `SIM_PORT` | HTTP server port | `8080` |
| `SIM_TICK_RATE_HZ` | Simulation tick rate (Hz) | `30` |
| `SIM_COMMAND_QUEUE_SIZE` | Command queue capacity | `100` |
| `SIM_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `SIM_LOG_FORMAT` | Log format (json/text) | `json` |

## Architecture

The system follows an **actor-based concurrent architecture**:

```
┌─────────────────────────────────────────────────┐
│              HTTP REST API                      │
└───────────────────┬─────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│           Command Queue (channel)               │
└───────────────────┬─────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│         Simulation Engine (Actor)               │
│  • Owns aircraft state (single-threaded)        │
│  • Processes commands from queue                │
│  • Publishes state updates                      │
└───────────────────┬─────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│         State Publisher (PubSub)                │
│  • Fan-out to multiple subscribers              │
│  • SSE streaming support                        │
└─────────────────────────────────────────────────┘
```

### Key Design Decisions

- **Actor Model**: Simulation engine is the sole owner of aircraft state
- **Channels**: All communication via channels (no shared mutable state)
- **Context**: Graceful shutdown via context cancellation
- **PubSub**: Fan-out pattern for state broadcasting

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architecture.

## Project Structure

```
.
├── cmd/
│   └── simulator/          # Main application entry point
├── internal/
│   ├── api/                # HTTP handlers and routing
│   │   ├── handlers/       # Endpoint handlers
│   │   ├── middleware/     # HTTP middleware
│   │   └── validation/     # Request validation
│   ├── simulator/          # Simulation engine
│   ├── pubsub/             # State publisher
│   ├── models/             # Data models
│   ├── environment/        # Environment effects
│   ├── config/             # Configuration
│   └── observability/      # Logging and metrics
├── pkg/
│   └── geo/                # Geographic utilities
├── configs/                # Configuration files
├── scripts/                # Utility scripts
├── tests/                  # Test suites
├── docs/                   # Documentation
│   ├── PRD.md              # Product requirements
│   ├── ARCHITECTURE.md     # Architecture design
│   ├── API.md              # API documentation
│   └── DESIGN_DECISIONS.md # Design rationale
└── README.md               # This file
```

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[PRD.md](docs/PRD.md)** - Product Requirements Document
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Architecture Design
- **[API.md](docs/API.md)** - REST API Reference
- **[DESIGN_DECISIONS.md](docs/DESIGN_DECISIONS.md)** - Design Rationale

## Development

### Code Style

Follow standard Go conventions:

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Vet code
go vet ./...
```

### Adding New Features

1. Update documentation first (docs/PRD.md)
2. Write tests (TDD approach)
3. Implement feature
4. Verify with race detector
5. Update API documentation if needed

## Assumptions & Tradeoffs

### Assumptions

- Single aircraft simulation (not multiple)
- Abstract flight model (simplified physics)
- Earth treated as sphere (not ellipsoid)
- Stateless design (no persistence)
- Single-process deployment

See [docs/DESIGN_DECISIONS.md](docs/DESIGN_DECISIONS.md) for complete list.

### Known Limitations

- No authentication/authorization
- No persistence (state lost on restart)
- Single aircraft only
- Simplified geographic calculations (~0.5% error)

### Future Improvements

- [ ] Prometheus metrics endpoint
- [ ] Multiple aircraft support
- [ ] Terrain collision detection (bonus)
- [ ] Command history API
- [ ] State persistence
- [ ] Web UI visualization
- [ ] Kubernetes deployment

## Performance

### Expected Performance

- **Tick Rate**: 30 Hz (configurable up to 100+ Hz)
- **CPU Usage**: ~3-5% on modern hardware
- **Memory**: ~20MB baseline + 50KB per SSE client
- **Latency**: <5ms for command submission, <100μs state query
- **Throughput**: 1000+ commands/sec

### Race Condition Testing

```bash
# Run with race detector (critical!)
go test -race -count=100 ./...

# Stress test
go test -race -parallel=20 -count=1000 ./internal/simulator
```

## Troubleshooting

### Server won't start

```bash
# Check port availability
lsof -i :8080

# Use different port
export SIM_PORT=8081
go run cmd/simulator/main.go
```

### Tests fail with race detector

```bash
# Race conditions detected - this is a critical error!
# Review the race detector output carefully
go test -race -v ./internal/simulator
```

### High CPU usage

```bash
# Reduce tick rate
export SIM_TICK_RATE_HZ=10
go run cmd/simulator/main.go
```

## Contributing

This is a home assignment project. Contributions are not expected, but feedback is welcome!

## License

MIT License - see LICENSE file for details.

## Author

**Meiron Tzhori**
- GitHub: [@meiron-tzhori](https://github.com/meiron-tzhori)
- Assignment: Senior Developer - Flight Simulator Backend

## Acknowledgments

- Assignment requirements provided by [Company Name]
- Built with Go 1.22
- Uses Gin web framework
- Inspired by real-world flight control systems

---

**Assignment Date**: February 2026  
**Status**: Implementation in progress  
**Documentation**: Complete
