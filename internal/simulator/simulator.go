package simulator

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/environment"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/pubsub"
	"github.com/meiron-tzhori/Flight-Simulator/pkg/geo"
)

// Simulator represents the aircraft simulation engine.
// It follows the Actor model: single goroutine owns all state.
type Simulator struct {
	// State (PRIVATE - only accessed in Run goroutine)
	state           models.AircraftState
	activeCommand   *models.Command
	trajectoryState *trajectoryState
	startTime       time.Time

	// Communication channels
	commandQueue  chan *models.Command
	stateRequests chan stateRequest

	// Components
	publisher   *pubsub.StatePublisher
	environment *environment.Environment

	// Configuration
	tickerInterval time.Duration
	config         config.SimulationConfig

	// Logger
	logger *slog.Logger
}

// stateRequest represents a request for current aircraft state.
type stateRequest struct {
	reply chan models.AircraftState
}

// trajectoryState tracks progress through a trajectory.
type trajectoryState struct {
	currentWaypointIndex int
}

// New creates a new simulator instance.
func New(cfg config.SimulationConfig, envCfg config.EnvironmentConfig, logger *slog.Logger) (*Simulator, error) {
	// Validate configuration
	if cfg.TickRateHz <= 0 {
		return nil, fmt.Errorf("tick rate must be positive")
	}

	// Calculate tick interval
	tickerInterval := time.Duration(float64(time.Second) / cfg.TickRateHz)

	// Initialize state
	initialState := models.AircraftState{
		Position: models.Position{
			Latitude:  cfg.InitialPosition.Latitude,
			Longitude: cfg.InitialPosition.Longitude,
			Altitude:  cfg.InitialPosition.Altitude,
		},
		Velocity: models.Velocity{
			GroundSpeed:   cfg.InitialVelocity.GroundSpeed,
			VerticalSpeed: cfg.InitialVelocity.VerticalSpeed,
		},
		Heading:   cfg.InitialHeading,
		Timestamp: time.Now(),
	}

	// Create environment
	env := environment.New(envCfg)

	s := &Simulator{
		state:           initialState,
		activeCommand:   nil,
		trajectoryState: nil,
		startTime:       time.Now(),
		commandQueue:    make(chan *models.Command, cfg.CommandQueueSize),
		stateRequests:   make(chan stateRequest),
		publisher:       pubsub.NewStatePublisher(10), // 10-item buffer per subscriber
		environment:     env,
		tickerInterval:  tickerInterval,
		config:          cfg,
		logger:          logger,
	}

	logger.Info("Simulator initialized",
		"tick_interval", tickerInterval,
		"initial_position", initialState.Position,
		"environment_enabled", env != nil && env.IsEnabled(),
	)

	if env != nil && env.IsEnabled() {
		if wind := env.GetWind(); wind != nil {
			logger.Info("Wind effect enabled",
				"direction", wind.GetVector().Direction,
				"speed_ms", wind.GetVector().Speed,
			)
		}
	}

	return s, nil
}

// Run starts the simulation loop. This is the Actor's main goroutine.
func (s *Simulator) Run(ctx context.Context) error {
	s.logger.Info("Starting simulation loop")

	ticker := time.NewTicker(s.tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Simulation loop shutting down")
			return ctx.Err()

		case <-ticker.C:
			s.tick()

		case cmd := <-s.commandQueue:
			s.handleCommand(cmd)

		case req := <-s.stateRequests:
			// Synchronous state query
			req.reply <- s.state
		}
	}
}

// SubmitCommand submits a command to the simulator.
func (s *Simulator) SubmitCommand(ctx context.Context, cmd *models.Command) error {
	select {
	case s.commandQueue <- cmd:
		s.logger.Debug("Command queued", "command_id", cmd.ID, "type", cmd.Type)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return models.ErrCommandQueueFull
	}
}

// GetState returns the current aircraft state.
func (s *Simulator) GetState(ctx context.Context) (models.AircraftState, error) {
	req := stateRequest{
		reply: make(chan models.AircraftState, 1),
	}

	select {
	case s.stateRequests <- req:
		state := <-req.reply
		return state, nil
	case <-ctx.Done():
		return models.AircraftState{}, ctx.Err()
	case <-time.After(1 * time.Second):
		return models.AircraftState{}, models.ErrTimeout
	}
}

// GetPublisher returns the state publisher for SSE subscriptions.
func (s *Simulator) GetPublisher() *pubsub.StatePublisher {
	return s.publisher
}

// tick performs one simulation step.
func (s *Simulator) tick() {
	// Calculate time since last tick
	deltaTime := s.tickerInterval.Seconds()

	// Apply environment effects if enabled
	effectiveVelocity := s.state.Velocity
	if s.environment != nil && s.environment.IsEnabled() {
		effectiveVelocity = s.environment.ApplyEffects(s.state.Heading, s.state.Velocity)
	}

	// Execute active command if present
	if s.activeCommand != nil {
		switch s.activeCommand.Type {
		case models.CommandTypeGoTo:
			s.executeGoTo(s.activeCommand.GoTo, deltaTime)
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

	// Add environment state to aircraft state
	if s.environment != nil {
		s.state.Environment = s.environment.GetState()
	}

	// Update timestamp
	s.state.Timestamp = time.Now()

	// Publish state to subscribers
	s.publisher.Publish(s.state)
}

// handleCommand processes a newly received command.
func (s *Simulator) handleCommand(cmd *models.Command) {
	s.logger.Info("Command received", "command_id", cmd.ID, "type", cmd.Type)

	// Reset trajectory state if switching from trajectory command
	if s.activeCommand != nil && s.activeCommand.Type == models.CommandTypeTrajectory {
		if cmd.Type != models.CommandTypeTrajectory {
			s.trajectoryState = nil
		}
	}

	// Store as active command
	s.activeCommand = cmd

	// Reset trajectory state for new trajectory commands
	if cmd.Type == models.CommandTypeTrajectory {
		s.trajectoryState = &trajectoryState{currentWaypointIndex: 0}
	}
}

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

// executeGoTo executes a go-to command.
func (s *Simulator) executeGoTo(cmd *models.GoToCommand, deltaTime float64) {
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
	s.executeGoTo(gotoCmd, deltaTime)
}

// executeHold executes a hold command (orbit at current position).
func (s *Simulator) executeHold(deltaTime float64, velocity models.Velocity) {
	// Simple hold: reduce speed to near-zero and stop climbing
	s.adjustSpeed(0, deltaTime)
	s.state.Velocity.VerticalSpeed = 0

	// Optional: Implement circular orbit pattern
	// For simplicity, just hover in place
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

	// Ensure speed doesn't go negative
	if s.state.Velocity.GroundSpeed < 0 {
		s.state.Velocity.GroundSpeed = 0
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
