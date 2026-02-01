# Implementation Guide

## Table of Contents

1. [Overview](#overview)
2. [Phase 1: Core Simulation Logic](#phase-1-core-simulation-logic)
3. [Phase 2: Command Handling](#phase-2-command-handling)
4. [Phase 3: API Implementation](#phase-3-api-implementation)
5. [Phase 4: Environment Effects](#phase-4-environment-effects)
6. [Phase 5: Testing](#phase-5-testing)
7. [Phase 6: Bonus Features](#phase-6-bonus-features)
8. [Verification Checklist](#verification-checklist)
9. [Common Pitfalls](#common-pitfalls)
10. [Debugging Tips](#debugging-tips)

---

## Overview

This guide provides step-by-step instructions for implementing the Flight Simulator backend. The skeleton code is already in place; this guide shows you how to fill in the `TODO` markers.

### Implementation Order

```
Phase 1: Core Simulation (3-4 hours)
  ↓
Phase 2: Command Handling (2-3 hours)
  ↓
Phase 3: API Implementation (2-3 hours)
  ↓
Phase 4: Environment Effects (1-2 hours)
  ↓
Phase 5: Testing (2-3 hours)
  ↓
Phase 6: Bonus Features (1-2 hours)
```

**Total Estimated Time**: 12-18 hours

---

## Phase 1: Core Simulation Logic

### 1.1 Implement Tick Method

**File**: `internal/simulator/simulator.go`

**Goal**: Update aircraft position based on velocity and heading.

```go
// tick performs one simulation step.
func (s *Simulator) tick() {
	// Calculate time since last tick
	deltaTime := s.tickerInterval.Seconds()

	// Apply environment effects if enabled
	effectiveVelocity := s.state.Velocity
	// TODO Phase 4: Add environment effects

	// Execute active command if present
	if s.activeCommand != nil {
		switch s.activeCommand.Type {
		case models.CommandTypeGoTo:
			s.executeGoTo(s.activeCommand.GoTo, deltaTime, effectiveVelocity)
		case models.CommandTypeTrajectory:
			s.executeTrajectory(s.activeCommand.Trajectory, deltaTime, effectiveVelocity)
		case models.CommandTypeHold:
			s.executeHold(deltaTime, effectiveVelocity)
		case models.CommandTypeStop:
			// Aircraft is stopped, no movement
		}
	} else {
		// No active command - maintain current heading and speed
		s.updatePosition(deltaTime, effectiveVelocity)
	}

	// Update timestamp
	s.state.Timestamp = time.Now()

	// Publish state to subscribers
	s.publisher.Publish(s.state)
}
```

### 1.2 Implement Position Update

**Add to**: `internal/simulator/simulator.go`

```go
// updatePosition updates aircraft position based on current velocity and heading.
func (s *Simulator) updatePosition(deltaTime float64, velocity models.Velocity) {
	// Calculate distance traveled
	distance := velocity.GroundSpeed * deltaTime

	// Convert heading to radians
	headingRad := s.state.Heading * math.Pi / 180.0

	// Calculate position change
	// Note: Simplified calculation treating Earth as a sphere
	const metersPerDegreeLat = 111000.0 // Approximate
	metersPerDegreeLon := 111000.0 * math.Cos(s.state.Position.Latitude*math.Pi/180.0)

	// North/South movement (latitude)
	deltaLat := (distance * math.Cos(headingRad)) / metersPerDegreeLat
	// East/West movement (longitude)
	deltaLon := (distance * math.Sin(headingRad)) / metersPerDegreeLon

	// Update position
	s.state.Position.Latitude += deltaLat
	s.state.Position.Longitude += deltaLon

	// Update altitude
	deltaAlt := velocity.VerticalSpeed * deltaTime
	s.state.Position.Altitude += deltaAlt

	// Ensure altitude doesn't go negative
	if s.state.Position.Altitude < 0 {
		s.state.Position.Altitude = 0
		s.state.Velocity.VerticalSpeed = 0
	}
}
```

### 1.3 Implement Go-To Logic

**Add to**: `internal/simulator/simulator.go`

```go
import (
	"github.com/meiron-tzhori/Flight-Simulator/pkg/geo"
	"math"
)

// executeGoTo executes a go-to command.
func (s *Simulator) executeGoTo(cmd *models.GoToCommand, deltaTime float64, velocity models.Velocity) {
	// Calculate distance to target
	distance := geo.Haversine(
		s.state.Position.Latitude,
		s.state.Position.Longitude,
		cmd.Target.Latitude,
		cmd.Target.Longitude,
	)

	// Check if target reached
	if distance < s.config.PositionTolerance {
		s.logger.Info("Target reached", "command_id", s.activeCommand.ID)
		s.activeCommand = nil // Command complete
		s.state.Velocity.GroundSpeed = 0
		s.state.Velocity.VerticalSpeed = 0
		return
	}

	// Calculate desired heading to target
	targetHeading := geo.Bearing(
		s.state.Position.Latitude,
		s.state.Position.Longitude,
		cmd.Target.Latitude,
		cmd.Target.Longitude,
	)

	// Adjust heading towards target (with turn rate limit)
	s.adjustHeading(targetHeading, deltaTime)

	// Set speed
	targetSpeed := s.config.DefaultSpeed
	if cmd.Speed != nil {
		targetSpeed = *cmd.Speed
	}
	s.adjustSpeed(targetSpeed, deltaTime)

	// Calculate target vertical speed for altitude change
	altitudeDiff := cmd.Target.Altitude - s.state.Position.Altitude
	timeToTarget := distance / s.state.Velocity.GroundSpeed
	if timeToTarget > 0 {
		desiredVerticalSpeed := altitudeDiff / timeToTarget
		// Clamp to max rates
		desiredVerticalSpeed = clamp(desiredVerticalSpeed, -s.config.MaxDescentRate, s.config.MaxClimbRate)
		s.state.Velocity.VerticalSpeed = desiredVerticalSpeed
	}

	// Update position
	s.updatePosition(deltaTime, s.state.Velocity)
}

// adjustHeading smoothly adjusts heading towards target.
func (s *Simulator) adjustHeading(targetHeading, deltaTime float64) {
	currentHeading := s.state.Heading

	// Calculate shortest angular distance
	diff := targetHeading - currentHeading
	if diff > 180 {
		diff -= 360
	} else if diff < -180 {
		diff += 360
	}

	// Apply turn rate limit
	maxTurn := s.config.HeadingChangeRate * deltaTime
	if math.Abs(diff) < maxTurn {
		s.state.Heading = targetHeading
	} else if diff > 0 {
		s.state.Heading += maxTurn
	} else {
		s.state.Heading -= maxTurn
	}

	// Normalize to 0-360
	s.state.Heading = math.Mod(s.state.Heading+360, 360)
}

// adjustSpeed smoothly adjusts speed towards target.
func (s *Simulator) adjustSpeed(targetSpeed, deltaTime float64) {
	currentSpeed := s.state.Velocity.GroundSpeed
	diff := targetSpeed - currentSpeed

	// Apply acceleration limit
	maxChange := s.config.SpeedChangeRate * deltaTime
	if math.Abs(diff) < maxChange {
		s.state.Velocity.GroundSpeed = targetSpeed
	} else if diff > 0 {
		s.state.Velocity.GroundSpeed += maxChange
	} else {
		s.state.Velocity.GroundSpeed -= maxChange
	}

	// Clamp to max speed
	if s.state.Velocity.GroundSpeed > s.config.MaxSpeed {
		s.state.Velocity.GroundSpeed = s.config.MaxSpeed
	}
}

// clamp clamps a value between min and max.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
```

### 1.4 Implement Trajectory Logic

**Add to**: `internal/simulator/simulator.go`

```go
// trajectoryState tracks progress through a trajectory.
type trajectoryState struct {
	currentWaypointIndex int
}

// Add to Simulator struct:
// trajectoryState *trajectoryState

// executeTrajectory executes a trajectory command.
func (s *Simulator) executeTrajectory(cmd *models.TrajectoryCommand, deltaTime float64, velocity models.Velocity) {
	// Initialize trajectory state if needed
	if s.trajectoryState == nil {
		s.trajectoryState = &trajectoryState{currentWaypointIndex: 0}
	}

	// Check if trajectory complete
	if s.trajectoryState.currentWaypointIndex >= len(cmd.Waypoints) {
		if cmd.Loop {
			// Restart from beginning
			s.trajectoryState.currentWaypointIndex = 0
			s.logger.Info("Trajectory looping", "command_id", s.activeCommand.ID)
		} else {
			// Trajectory complete
			s.logger.Info("Trajectory complete", "command_id", s.activeCommand.ID)
			s.activeCommand = nil
			s.trajectoryState = nil
			s.state.Velocity.GroundSpeed = 0
			s.state.Velocity.VerticalSpeed = 0
			return
		}
	}

	// Get current waypoint
	waypoint := cmd.Waypoints[s.trajectoryState.currentWaypointIndex]

	// Create a temporary go-to command for current waypoint
	gotoCmd := &models.GoToCommand{
		Target: waypoint.Position,
		Speed:  waypoint.Speed,
	}

	// Calculate distance to waypoint
	distance := geo.Haversine(
		s.state.Position.Latitude,
		s.state.Position.Longitude,
		waypoint.Position.Latitude,
		waypoint.Position.Longitude,
	)

	// Check if waypoint reached
	if distance < s.config.PositionTolerance {
		s.logger.Info("Waypoint reached",
			"command_id", s.activeCommand.ID,
			"waypoint_index", s.trajectoryState.currentWaypointIndex,
		)
		s.trajectoryState.currentWaypointIndex++
		return
	}

	// Execute go-to logic for current waypoint
	s.executeGoTo(gotoCmd, deltaTime, velocity)
}
```

### 1.5 Implement Stop and Hold

**Add to**: `internal/simulator/simulator.go`

```go
// executeHold executes a hold command (orbit at current position).
func (s *Simulator) executeHold(deltaTime float64, velocity models.Velocity) {
	// Simple hold: reduce speed to near-zero and stop climbing
	s.adjustSpeed(0, deltaTime)
	s.state.Velocity.VerticalSpeed = 0

	// Optional: Implement circular orbit pattern
	// For simplicity, just hover in place
	s.updatePosition(deltaTime, s.state.Velocity)
}
```

---

## Phase 2: Command Handling

### 2.1 Implement Validation

**Create file**: `internal/api/validation/validate.go`

```go
package validation

import (
	"fmt"

	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

// ValidatePosition validates geographic coordinates.
func ValidatePosition(pos models.Position) error {
	if pos.Latitude < -90 || pos.Latitude > 90 {
		return fmt.Errorf("%w: %f", models.ErrInvalidLatitude, pos.Latitude)
	}
	if pos.Longitude < -180 || pos.Longitude > 180 {
		return fmt.Errorf("%w: %f", models.ErrInvalidLongitude, pos.Longitude)
	}
	if pos.Altitude < 0 {
		return fmt.Errorf("%w: %f", models.ErrInvalidAltitude, pos.Altitude)
	}
	return nil
}

// ValidateSpeed validates speed value.
func ValidateSpeed(speed float64, maxSpeed float64) error {
	if speed < 0 {
		return models.ErrInvalidSpeed
	}
	if speed > maxSpeed {
		return fmt.Errorf("%w: %f > %f", models.ErrSpeedExceedsMax, speed, maxSpeed)
	}
	return nil
}

// ValidateGoToCommand validates a go-to command.
func ValidateGoToCommand(cmd *models.GoToCommand, maxSpeed float64) error {
	if err := ValidatePosition(cmd.Target); err != nil {
		return err
	}
	if cmd.Speed != nil {
		if err := ValidateSpeed(*cmd.Speed, maxSpeed); err != nil {
			return err
		}
	}
	return nil
}

// ValidateTrajectoryCommand validates a trajectory command.
func ValidateTrajectoryCommand(cmd *models.TrajectoryCommand, maxSpeed float64) error {
	if len(cmd.Waypoints) == 0 {
		return models.ErrEmptyWaypoints
	}

	for i, wp := range cmd.Waypoints {
		if err := ValidatePosition(wp.Position); err != nil {
			return fmt.Errorf("waypoint %d: %w", i, err)
		}
		if wp.Speed != nil {
			if err := ValidateSpeed(*wp.Speed, maxSpeed); err != nil {
				return fmt.Errorf("waypoint %d: %w", i, err)
			}
		}
	}

	return nil
}
```

### 2.2 Implement Command Handlers

**Update**: `internal/api/handlers/command.go`

```go
package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meiron-tzhori/Flight-Simulator/internal/api/validation"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

// GoToRequest represents the request body for go-to command.
type GoToRequest struct {
	Lat   float64  `json:"lat" binding:"required"`
	Lon   float64  `json:"lon" binding:"required"`
	Alt   float64  `json:"alt" binding:"required"`
	Speed *float64 `json:"speed,omitempty"`
}

// GoTo handles POST /command/goto
func (h *CommandHandler) GoTo(c *gin.Context) {
	var req GoToRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Create command
	cmd := models.NewCommand(models.CommandTypeGoTo)
	cmd.GoTo = &models.GoToCommand{
		Target: models.Position{
			Latitude:  req.Lat,
			Longitude: req.Lon,
			Altitude:  req.Alt,
		},
		Speed: req.Speed,
	}

	// Validate
	if err := validation.ValidateGoToCommand(cmd.GoTo, 250.0); err != nil {
		h.logger.Warn("Validation failed", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    getErrorCode(err),
				Message: err.Error(),
			},
		})
		return
	}

	// Submit to simulator
	if err := h.simulator.SubmitCommand(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to submit command", "error", err)
		if errors.Is(err, models.ErrCommandQueueFull) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "QUEUE_FULL",
					Message: "Command queue is full, please retry",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to submit command",
				},
			})
		}
		return
	}

	// Success
	c.JSON(http.StatusOK, models.CommandResponse{
		Status:    "accepted",
		CommandID: cmd.ID,
		Message:   "Go-to command accepted",
		Target:    &cmd.GoTo.Target,
	})
}

// TrajectoryRequest represents the request body for trajectory command.
type TrajectoryRequest struct {
	Waypoints []WaypointRequest `json:"waypoints" binding:"required,min=1"`
	Loop      bool              `json:"loop"`
}

type WaypointRequest struct {
	Lat   float64  `json:"lat" binding:"required"`
	Lon   float64  `json:"lon" binding:"required"`
	Alt   float64  `json:"alt" binding:"required"`
	Speed *float64 `json:"speed,omitempty"`
}

// Trajectory handles POST /command/trajectory
func (h *CommandHandler) Trajectory(c *gin.Context) {
	var req TrajectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Create command
	cmd := models.NewCommand(models.CommandTypeTrajectory)
	waypoints := make([]models.Waypoint, len(req.Waypoints))
	for i, wp := range req.Waypoints {
		waypoints[i] = models.Waypoint{
			Position: models.Position{
				Latitude:  wp.Lat,
				Longitude: wp.Lon,
				Altitude:  wp.Alt,
			},
			Speed: wp.Speed,
		}
	}
	cmd.Trajectory = &models.TrajectoryCommand{
		Waypoints: waypoints,
		Loop:      req.Loop,
	}

	// Validate
	if err := validation.ValidateTrajectoryCommand(cmd.Trajectory, 250.0); err != nil {
		h.logger.Warn("Validation failed", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    getErrorCode(err),
				Message: err.Error(),
			},
		})
		return
	}

	// Submit to simulator
	if err := h.simulator.SubmitCommand(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to submit command", "error", err)
		if errors.Is(err, models.ErrCommandQueueFull) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "QUEUE_FULL",
					Message: "Command queue is full, please retry",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to submit command",
				},
			})
		}
		return
	}

	// Success
	c.JSON(http.StatusOK, models.CommandResponse{
		Status:        "accepted",
		CommandID:     cmd.ID,
		Message:       "Trajectory command accepted",
		WaypointCount: len(waypoints),
	})
}

// Stop handles POST /command/stop
func (h *CommandHandler) Stop(c *gin.Context) {
	cmd := models.NewCommand(models.CommandTypeStop)

	if err := h.simulator.SubmitCommand(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to submit command", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to submit stop command",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.CommandResponse{
		Status:    "accepted",
		CommandID: cmd.ID,
		Message:   "Stop command accepted",
	})
}

// Hold handles POST /command/hold
func (h *CommandHandler) Hold(c *gin.Context) {
	cmd := models.NewCommand(models.CommandTypeHold)

	if err := h.simulator.SubmitCommand(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to submit command", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to submit hold command",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.CommandResponse{
		Status:    "accepted",
		CommandID: cmd.ID,
		Message:   "Hold command accepted",
	})
}

// getErrorCode extracts error code from error.
func getErrorCode(err error) string {
	switch {
	case errors.Is(err, models.ErrInvalidLatitude):
		return "INVALID_LATITUDE"
	case errors.Is(err, models.ErrInvalidLongitude):
		return "INVALID_LONGITUDE"
	case errors.Is(err, models.ErrInvalidAltitude):
		return "INVALID_ALTITUDE"
	case errors.Is(err, models.ErrInvalidSpeed):
		return "INVALID_SPEED"
	case errors.Is(err, models.ErrEmptyWaypoints):
		return "EMPTY_WAYPOINTS"
	case errors.Is(err, models.ErrSpeedExceedsMax):
		return "SPEED_EXCEEDS_MAX"
	default:
		return "VALIDATION_ERROR"
	}
}
```

---

## Phase 3: API Implementation

### 3.1 Implement SSE Streaming

**Update**: `internal/api/handlers/stream.go`

```go
package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

// Stream handles GET /stream
func (h *StreamHandler) Stream(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Get flusher
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		h.logger.Error("Streaming not supported")
		c.String(http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// Subscribe to state updates
	subID := uuid.New().String()
	publisher := h.simulator.GetPublisher()
	stateChan := publisher.Subscribe(subID)
	defer publisher.Unsubscribe(subID)

	h.logger.Info("SSE client connected", "subscriber_id", subID)

	// Throttle to 10 Hz (every 100ms)
	throttle := time.NewTicker(100 * time.Millisecond)
	defer throttle.Stop()

	var latestState *models.AircraftState

	for {
		select {
		case state, ok := <-stateChan:
			if !ok {
				// Channel closed
				h.logger.Info("State channel closed", "subscriber_id", subID)
				return
			}
			// Cache latest state
			latestState = &state

		case <-throttle.C:
			if latestState != nil {
				// Send SSE event
				data, err := json.Marshal(latestState)
				if err != nil {
					h.logger.Error("Failed to marshal state", "error", err)
					continue
				}

				// Write SSE format: "data: {json}\n\n"
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
				flusher.Flush()
			}

		case <-c.Request.Context().Done():
			// Client disconnected
			h.logger.Info("SSE client disconnected", "subscriber_id", subID)
			return
		}
	}
}
```

**Add import**:
```go
import (
	"net/http"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)
```

### 3.2 Update Health Handler

**Update**: `internal/api/handlers/health.go`

```go
// Add to HealthHandler struct:
type HealthHandler struct {
	simulator *simulator.Simulator
	logger    *slog.Logger
	startTime time.Time
	tickRate  float64  // Add this
}

// Update NewHealthHandler:
func NewHealthHandler(sim *simulator.Simulator, logger *slog.Logger, tickRate float64) *HealthHandler {
	return &HealthHandler{
		simulator: sim,
		logger:    logger,
		startTime: time.Now(),
		tickRate:  tickRate,
	}
}

// Update Health method:
func (h *HealthHandler) Health(c *gin.Context) {
	response := models.HealthResponse{
		Status:            "healthy",
		SimulationRunning: true, // TODO: Add actual check
		TickRateHz:        h.tickRate,
		UptimeSeconds:     time.Since(h.startTime).Seconds(),
		Timestamp:         time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
```

**Update server.go** to pass tickRate:
```go
healthHandler := handlers.NewHealthHandler(sim, logger, cfg.Simulation.TickRateHz)
```

---

## Phase 4: Environment Effects

### 4.1 Implement Wind Effect

**Create file**: `internal/environment/wind.go`

```go
package environment

import (
	"math"

	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

// WindEffect applies wind to aircraft velocity.
type WindEffect struct {
	direction float64 // degrees
	speed     float64 // m/s
}

// NewWindEffect creates a new wind effect.
func NewWindEffect(direction, speed float64) *WindEffect {
	return &WindEffect{
		direction: direction,
		speed:     speed,
	}
}

// Apply applies wind effect to velocity.
func (w *WindEffect) Apply(heading float64, velocity models.Velocity) models.Velocity {
	// Convert to radians
	headingRad := heading * math.Pi / 180.0
	windDirRad := w.direction * math.Pi / 180.0

	// Calculate aircraft velocity components
	acNorth := velocity.GroundSpeed * math.Cos(headingRad)
	acEast := velocity.GroundSpeed * math.Sin(headingRad)

	// Calculate wind velocity components
	windNorth := w.speed * math.Cos(windDirRad)
	windEast := w.speed * math.Sin(windDirRad)

	// Add wind effect
	totalNorth := acNorth + windNorth
	totalEast := acEast + windEast

	// Calculate new ground speed (with wind effect)
	newGroundSpeed := math.Sqrt(totalNorth*totalNorth + totalEast*totalEast)

	return models.Velocity{
		GroundSpeed:   newGroundSpeed,
		VerticalSpeed: velocity.VerticalSpeed,
	}
}
```

### 4.2 Create Environment Manager

**Create file**: `internal/environment/environment.go`

```go
package environment

import (
	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

// Environment manages environmental effects.
type Environment struct {
	wind *WindEffect
	// TODO: Add humidity, terrain
}

// New creates a new environment from configuration.
func New(cfg config.EnvironmentConfig) *Environment {
	if !cfg.Enabled {
		return nil
	}

	var wind *WindEffect
	if cfg.Wind.Enabled {
		wind = NewWindEffect(cfg.Wind.Direction, cfg.Wind.Speed)
	}

	return &Environment{
		wind: wind,
	}
}

// ApplyEffects applies all enabled environmental effects.
func (e *Environment) ApplyEffects(heading float64, velocity models.Velocity) models.Velocity {
	if e == nil {
		return velocity
	}

	result := velocity

	if e.wind != nil {
		result = e.wind.Apply(heading, result)
	}

	// TODO: Apply other effects

	return result
}

// GetState returns environment state for API responses.
func (e *Environment) GetState() *models.EnvironmentState {
	if e == nil {
		return nil
	}

	state := &models.EnvironmentState{}

	if e.wind != nil {
		state.Wind = &models.WindVector{
			Direction: e.wind.direction,
			Speed:     e.wind.speed,
		}
	}

	return state
}
```

### 4.3 Integrate Environment into Simulator

**Update**: `internal/simulator/simulator.go`

```go
import (
	"github.com/meiron-tzhori/Flight-Simulator/internal/environment"
)

// Add to Simulator struct:
type Simulator struct {
	// ... existing fields ...
	environment *environment.Environment
}

// Update New function:
func New(cfg config.SimulationConfig, envCfg config.EnvironmentConfig, logger *slog.Logger) (*Simulator, error) {
	// ... existing code ...

	// Create environment
	env := environment.New(envCfg)

	s := &Simulator{
		// ... existing fields ...
		environment: env,
	}

	return s, nil
}

// Update tick method:
func (s *Simulator) tick() {
	deltaTime := s.tickerInterval.Seconds()

	// Apply environment effects
	effectiveVelocity := s.state.Velocity
	if s.environment != nil {
		effectiveVelocity = s.environment.ApplyEffects(s.state.Heading, s.state.Velocity)
	}

	// ... rest of tick logic ...

	// Add environment to state
	if s.environment != nil {
		s.state.Environment = s.environment.GetState()
	}

	s.state.Timestamp = time.Now()
	s.publisher.Publish(s.state)
}
```

**Update main.go**:
```go
sim, err := simulator.New(cfg.Simulation, cfg.Environment, logger)
```

---

## Phase 5: Testing

### 5.1 Unit Test Example - Geo Package

**Create file**: `pkg/geo/distance_test.go`

```go
package geo

import (
	"math"
	"testing"
)

func TestHaversine(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64 // meters
		tolerance float64
	}{
		{
			name:      "same point",
			lat1:      32.0853,
			lon1:      34.7818,
			lat2:      32.0853,
			lon2:      34.7818,
			expected:  0,
			tolerance: 1,
		},
		{
			name:      "Tel Aviv to Jerusalem",
			lat1:      32.0853,
			lon1:      34.7818,
			lat2:      31.7683,
			lon2:      35.2137,
			expected:  54000, // ~54km
			tolerance: 1000,  // 1km tolerance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Haversine(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			diff := math.Abs(got - tt.expected)
			if diff > tt.tolerance {
				t.Errorf("Haversine() = %v, want %v (±%v)",
					got, tt.expected, tt.tolerance)
			}
		})
	}
}
```

### 5.2 Unit Test Example - Validation

**Create file**: `internal/api/validation/validate_test.go`

```go
package validation

import (
	"testing"

	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

func TestValidatePosition(t *testing.T) {
	tests := []struct {
		name    string
		pos     models.Position
		wantErr bool
	}{
		{
			name: "valid position",
			pos: models.Position{
				Latitude:  32.0853,
				Longitude: 34.7818,
				Altitude:  1000.0,
			},
			wantErr: false,
		},
		{
			name: "latitude too high",
			pos: models.Position{
				Latitude:  91.0,
				Longitude: 34.7818,
				Altitude:  1000.0,
			},
			wantErr: true,
		},
		{
			name: "negative altitude",
			pos: models.Position{
				Latitude:  32.0853,
				Longitude: 34.7818,
				Altitude:  -100.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePosition(tt.pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePosition() error = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}
```

### 5.3 Integration Test Example

**Create file**: `tests/integration/api_test.go`

```go
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/meiron-tzhori/Flight-Simulator/internal/api"
	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/observability"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

func TestGoToCommand(t *testing.T) {
	// Setup
	cfg := getTestConfig()
	logger := observability.NewLogger(cfg.Logging)
	sim, _ := simulator.New(cfg.Simulation, cfg.Environment, logger)
	server := api.NewServer(cfg.Server, sim, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start simulator
	go sim.Run(ctx)
	time.Sleep(100 * time.Millisecond)

	// Create test server
	ts := httptest.NewServer(server.Handler)
	defer ts.Close()

	// Send goto command
	payload := `{"lat": 32.0853, "lon": 34.7818, "alt": 1000.0}`
	resp, err := http.Post(
		ts.URL+"/command/goto",
		"application/json",
		strings.NewReader(payload),
	)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result models.CommandResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Status != "accepted" {
		t.Errorf("Expected status 'accepted', got '%s'", result.Status)
	}
}

func getTestConfig() *config.Config {
	// Return minimal test config
	return &config.Config{
		// ... minimal config for testing ...
	}
}
```

---

## Phase 6: Bonus Features

### 6.1 Add Active Command Info to State

**Update**: `internal/simulator/simulator.go`

```go
// In tick() method, add:
if s.activeCommand != nil {
	s.state.ActiveCommand = &models.CommandInfo{
		Type: string(s.activeCommand.Type),
	}

	// Add target info for goto/trajectory
	if s.activeCommand.Type == models.CommandTypeGoTo {
		s.state.ActiveCommand.Target = &s.activeCommand.GoTo.Target
		// Calculate ETA
		dist := geo.Haversine(
			s.state.Position.Latitude,
			s.state.Position.Longitude,
			s.activeCommand.GoTo.Target.Latitude,
			s.activeCommand.GoTo.Target.Longitude,
		)
		if s.state.Velocity.GroundSpeed > 0 {
			s.state.ActiveCommand.ETASeconds = dist / s.state.Velocity.GroundSpeed
		}
	}
} else {
	s.state.ActiveCommand = nil
}
```

---

## Verification Checklist

### Code Quality
- [ ] All code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] All imports organized
- [ ] No unused variables or imports

### Functionality
- [ ] Simulator starts without errors
- [ ] Health endpoint returns 200
- [ ] State endpoint returns current position
- [ ] Go-to command moves aircraft
- [ ] Trajectory follows waypoints
- [ ] Stop command works
- [ ] Hold command works
- [ ] SSE streaming works

### Concurrency
- [ ] `go test -race ./...` passes with no warnings
- [ ] Stress test: `go test -race -count=100 ./internal/simulator`
- [ ] No deadlocks under load
- [ ] Graceful shutdown works (Ctrl+C)

### Validation
- [ ] Invalid coordinates rejected (400)
- [ ] Negative altitude rejected (400)
- [ ] Empty waypoints rejected (400)
- [ ] Queue full returns 503

### Documentation
- [ ] All public functions have comments
- [ ] README examples work
- [ ] API documentation is accurate
- [ ] curl-examples.sh runs successfully

---

## Common Pitfalls

### 1. Race Conditions

❌ **Wrong**: Sharing state without synchronization
```go
// DON'T DO THIS
type Simulator struct {
    state models.AircraftState  // shared!
    mu    sync.Mutex
}

func (s *Simulator) GetState() models.AircraftState {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.state  // Easy to forget locks!
}
```

✅ **Correct**: Actor model with channels
```go
// DO THIS
func (s *Simulator) GetState(ctx context.Context) (models.AircraftState, error) {
    req := stateRequest{reply: make(chan models.AircraftState, 1)}
    s.stateRequests <- req
    return <-req.reply, nil
}
```

### 2. Forgetting to Normalize Angles

❌ **Wrong**: Angles can go beyond 360°
```go
heading += turnRate * deltaTime  // Can become 450°!
```

✅ **Correct**: Always normalize
```go
heading = math.Mod(heading + turnRate * deltaTime + 360, 360)
```

### 3. Not Handling Context Cancellation

❌ **Wrong**: Ignoring context
```go
func (s *Simulator) Run(ctx context.Context) error {
    for {  // Runs forever!
        s.tick()
    }
}
```

✅ **Correct**: Check context
```go
func (s *Simulator) Run(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            s.tick()
        }
    }
}
```

### 4. Division by Zero

❌ **Wrong**: Not checking for zero
```go
eta := distance / speed  // Crash if speed is 0!
```

✅ **Correct**: Check first
```go
if speed > 0 {
    eta := distance / speed
}
```

### 5. Not Cleaning Up Goroutines

❌ **Wrong**: Goroutine leak
```go
go func() {
    for state := range stateChan {
        // Never exits if channel isn't closed!
    }
}()
```

✅ **Correct**: Context-aware
```go
go func() {
    for {
        select {
        case state := <-stateChan:
            // handle
        case <-ctx.Done():
            return
        }
    }
}()
```

---

## Debugging Tips

### Enable Debug Logging

```yaml
# config.yaml
logging:
  level: "debug"  # Change from "info"
```

### Check Race Conditions

```bash
# Run with race detector
go run -race cmd/simulator/main.go

# Test with race detector
go test -race ./...
```

### Inspect State

```bash
# Watch state continuously
watch -n 0.5 'curl -s localhost:8080/state | jq .'

# Just position
curl -s localhost:8080/state | jq '.position'
```

### Profile Performance

```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profile
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### Test Concurrency

```bash
# Parallel requests
for i in {1..100}; do
  curl -X POST localhost:8080/command/goto \
    -H 'Content-Type: application/json' \
    -d '{"lat":32.1,"lon":34.8,"alt":1000}' &
done
wait
```

---

## Next Steps After Implementation

1. **Run the Demo**
   ```bash
   make build
   make run
   # In another terminal:
   make demo
   ```

2. **Verify with Tests**
   ```bash
   make test-race
   make test-coverage
   ```

3. **Load Test**
   ```bash
   # Install hey: go install github.com/rakyll/hey@latest
   hey -n 1000 -c 10 http://localhost:8080/state
   ```

4. **Create Demo Video/GIF** (optional)
   - Show curl commands
   - Show SSE streaming
   - Show state changes

5. **Final Documentation Review**
   - Update README with any changes
   - Verify all curl examples work
   - Check that DESIGN_DECISIONS.md reflects reality

---

**Good luck with the implementation!** 🚀

Remember: **Focus on correctness first, then performance.** The race detector is your friend!
