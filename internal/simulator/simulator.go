package simulator

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/pubsub"
)

// Simulator represents the aircraft simulation engine.
// It follows the Actor model: single goroutine owns all state.
type Simulator struct {
	// State (PRIVATE - only accessed in Run goroutine)
	state         models.AircraftState
	activeCommand *models.Command
	startTime     time.Time

	// Communication channels
	commandQueue  chan *models.Command
	stateRequests chan stateRequest

	// Components
	publisher *pubsub.StatePublisher

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

// New creates a new simulator instance.
func New(cfg config.SimulationConfig, logger *slog.Logger) (*Simulator, error) {
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

	s := &Simulator{
		state:          initialState,
		activeCommand:  nil,
		startTime:      time.Now(),
		commandQueue:   make(chan *models.Command, cfg.CommandQueueSize),
		stateRequests:  make(chan stateRequest),
		publisher:      pubsub.NewStatePublisher(10), // 10-item buffer per subscriber
		tickerInterval: tickerInterval,
		config:         cfg,
		logger:         logger,
	}

	logger.Info("Simulator initialized",
		"tick_interval", tickerInterval,
		"initial_position", initialState.Position,
	)

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
			// TODO: Implement tick logic
			s.tick()

		case cmd := <-s.commandQueue:
			// TODO: Implement command handling
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
	// TODO: Implement simulation tick
	// 1. Calculate deltaTime
	// 2. Execute active command logic
	// 3. Update position based on velocity and heading
	// 4. Update timestamp
	// 5. Publish state

	s.state.Timestamp = time.Now()
	s.publisher.Publish(s.state)
}

// handleCommand processes a newly received command.
func (s *Simulator) handleCommand(cmd *models.Command) {
	// TODO: Implement command handling
	// 1. Validate command (if needed)
	// 2. Store as activeCommand
	// 3. Log acceptance

	s.logger.Info("Command received", "command_id", cmd.ID, "type", cmd.Type)
	s.activeCommand = cmd
}
