package simulator

import (
	"context"
	"testing"
	"time"

	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

func createTestConfig() *config.Config {
	return &config.Config{
		Initial: config.InitialState{
			Latitude:  32.0,
			Longitude: 34.0,
			Altitude:  1000.0,
			Heading:   0.0,
		},
		Aircraft: config.AircraftConfig{
			MaxSpeed:        250.0,
			MaxClimbRate:    15.0,
			MaxDescentRate:  10.0,
			TurnRate:        3.0,
			CruiseSpeed:     100.0,
			CruiseAltitude:  1500.0,
			MinSpeed:        30.0,
			MaxAcceleration: 5.0,
		},
		Environment: config.EnvironmentConfig{
			WindDirection: 0.0,
			WindSpeed:     0.0,
		},
		Simulation: config.SimulationConfig{
			UpdateInterval:   100 * time.Millisecond,
			CommandQueueSize: 10,
		},
	}
}

func TestSimulator_NewSimulator(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulator(cfg)

	if sim == nil {
		t.Fatal("NewSimulator() returned nil")
	}

	// Check initial state
	state := sim.GetState()

	if state.Position.Latitude != 32.0 {
		t.Errorf("Initial latitude = %f, want 32.0", state.Position.Latitude)
	}

	if state.Position.Longitude != 34.0 {
		t.Errorf("Initial longitude = %f, want 34.0", state.Position.Longitude)
	}

	if state.Position.Altitude != 1000.0 {
		t.Errorf("Initial altitude = %f, want 1000.0", state.Position.Altitude)
	}

	if state.Heading != 0.0 {
		t.Errorf("Initial heading = %f, want 0.0", state.Heading)
	}
}

func TestSimulator_GetState(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulator(cfg)

	state := sim.GetState()

	// State should have recent timestamp
	if time.Since(state.Timestamp) > 1*time.Second {
		t.Errorf("State timestamp is too old: %v", state.Timestamp)
	}

	// State should have valid values
	if state.Position.Latitude < -90 || state.Position.Latitude > 90 {
		t.Errorf("Invalid latitude: %f", state.Position.Latitude)
	}

	if state.Position.Longitude < -180 || state.Position.Longitude > 180 {
		t.Errorf("Invalid longitude: %f", state.Position.Longitude)
	}

	if state.Position.Altitude < 0 {
		t.Errorf("Invalid altitude: %f", state.Position.Altitude)
	}
}

func TestSimulator_SendCommand_GoTo(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulator(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond) // Let simulator start

	cmd := models.GoToCommand{
		Target: models.Position{
			Latitude:  32.1,
			Longitude: 34.1,
			Altitude:  1500.0,
		},
		Speed: ptr(100.0),
	}

	err := sim.SendCommand(cmd)
	if err != nil {
		t.Fatalf("SendCommand() error = %v", err)
	}

	// Wait for command to be processed
	time.Sleep(200 * time.Millisecond)

	// Check that aircraft is moving
	state := sim.GetState()
	if state.Velocity.GroundSpeed <= 0 {
		t.Error("Aircraft not moving after goto command")
	}
}

func TestSimulator_SendCommand_Trajectory(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulator(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	cmd := models.TrajectoryCommand{
		Waypoints: []models.Waypoint{
			{
				Position: models.Position{Latitude: 32.05, Longitude: 34.05, Altitude: 1200},
				Speed:    ptr(50.0),
			},
			{
				Position: models.Position{Latitude: 32.1, Longitude: 34.1, Altitude: 1500},
				Speed:    ptr(100.0),
			},
		},
		Loop: false,
	}

	err := sim.SendCommand(cmd)
	if err != nil {
		t.Fatalf("SendCommand() error = %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	state := sim.GetState()
	if state.Velocity.GroundSpeed <= 0 {
		t.Error("Aircraft not moving after trajectory command")
	}
}

func TestSimulator_SendCommand_Stop(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulator(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	// First send goto to get aircraft moving
	gotoCmd := models.GoToCommand{
		Target: models.Position{
			Latitude:  32.1,
			Longitude: 34.1,
			Altitude:  1000.0,
		},
		Speed: ptr(100.0),
	}

	if err := sim.SendCommand(gotoCmd); err != nil {
		t.Fatalf("SendCommand(goto) error = %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Now send stop
	stopCmd := models.StopCommand{}
	if err := sim.SendCommand(stopCmd); err != nil {
		t.Fatalf("SendCommand(stop) error = %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Aircraft should be stopped or stopping
	state := sim.GetState()
	// Speed should be reducing (may not be zero yet due to deceleration)
	if state.Velocity.GroundSpeed > 50 {
		t.Errorf("Aircraft speed still high after stop: %f m/s", state.Velocity.GroundSpeed)
	}
}

func TestSimulator_SendCommand_Hold(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulator(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	holdCmd := models.HoldCommand{}
	err := sim.SendCommand(holdCmd)
	if err != nil {
		t.Fatalf("SendCommand(hold) error = %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// In hold mode, aircraft should maintain altitude
	state := sim.GetState()
	if state.Position.Altitude < 900 || state.Position.Altitude > 1100 {
		t.Errorf("Aircraft altitude changed significantly in hold: %f", state.Position.Altitude)
	}
}

func TestSimulator_Subscribe(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulator(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	// Subscribe to state updates
	subCtx, subCancel := context.WithTimeout(ctx, 1*time.Second)
	defer subCancel()

	stateChan := sim.Subscribe(subCtx)

	// Should receive at least one state update
	select {
	case state := <-stateChan:
		if state.Position.Latitude != 32.0 {
			t.Errorf("Received state with unexpected latitude: %f", state.Position.Latitude)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Did not receive state update within timeout")
	}
}

func TestSimulator_MultipleSubscribers(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulator(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	// Create multiple subscribers
	subCtx, subCancel := context.WithTimeout(ctx, 1*time.Second)
	defer subCancel()

	chan1 := sim.Subscribe(subCtx)
	chan2 := sim.Subscribe(subCtx)
	chan3 := sim.Subscribe(subCtx)

	// All should receive updates
	receivedCount := 0
	for i := 0; i < 3; i++ {
		select {
		case <-chan1:
			receivedCount++
		case <-chan2:
			receivedCount++
		case <-chan3:
			receivedCount++
		case <-time.After(500 * time.Millisecond):
			t.Error("Not all subscribers received updates")
			return
		}
	}

	if receivedCount < 3 {
		t.Errorf("Expected 3 updates, got %d", receivedCount)
	}
}

func TestSimulator_Run_Context_Cancellation(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulator(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)
	go func() {
		sim.Run(ctx)
		done <- true
	}()

	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Simulator should stop
	select {
	case <-done:
		// Success - simulator stopped
	case <-time.After(2 * time.Second):
		t.Error("Simulator did not stop after context cancellation")
	}
}

func TestSimulator_CommandQueue_Full(t *testing.T) {
	cfg := createTestConfig()
	cfg.Simulation.CommandQueueSize = 2 // Small queue
	sim := NewSimulator(cfg)

	// Don't start simulator (commands won't be processed)

	// Fill the queue
	for i := 0; i < 2; i++ {
		cmd := models.GoToCommand{
			Target: models.Position{Latitude: 32.0, Longitude: 34.0, Altitude: 1000},
		}
		if err := sim.SendCommand(cmd); err != nil {
			t.Fatalf("Failed to send command %d: %v", i, err)
		}
	}

	// Next command should fail (queue full)
	cmd := models.GoToCommand{
		Target: models.Position{Latitude: 32.0, Longitude: 34.0, Altitude: 1000},
	}
	err := sim.SendCommand(cmd)
	if err == nil {
		t.Error("Expected error when queue is full, got nil")
	}
}

// Helper function
func ptr(f float64) *float64 {
	return &f
}
