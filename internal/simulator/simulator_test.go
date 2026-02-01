package simulator

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

func createTestConfig() (config.SimulationConfig, config.EnvironmentConfig) {
	simCfg := config.SimulationConfig{
		TickRateHz: 10.0,
		CommandQueueSize: 10,
		InitialPosition: config.PositionConfig{
			Latitude:  32.0,
			Longitude: 34.0,
			Altitude:  1000.0,
		},
		InitialVelocity: config.VelocityConfig{
			GroundSpeed:   0,
			VerticalSpeed: 0,
		},
		InitialHeading:    0.0,
		DefaultSpeed:      100.0,
		MaxSpeed:          250.0,
		MaxClimbRate:      15.0,
		MaxDescentRate:    10.0,
		PositionTolerance: 10.0,
		HeadingChangeRate: 30.0,
		SpeedChangeRate:   50.0,
	}
	
	envCfg := config.EnvironmentConfig{
		Enabled: false,
	}
	
	return simCfg, envCfg
}

func TestSimulator_New(t *testing.T) {
	simCfg, envCfg := createTestConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	
	sim, err := New(simCfg, envCfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	
	if sim == nil {
		t.Fatal("New() returned nil")
	}
	
	// Check initial state
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Start simulator to enable GetState
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	go sim.Run(ctx2)
	time.Sleep(50 * time.Millisecond)
	
	state, err := sim.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	
	if state.Position.Latitude != 32.0 {
		t.Errorf("Initial latitude = %f, want 32.0", state.Position.Latitude)
	}
}

func TestSimulator_GetState(t *testing.T) {
	simCfg, envCfg := createTestConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	
	sim, err := New(simCfg, envCfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond)
	
	getCtx, getCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer getCancel()
	
	state, err := sim.GetState(getCtx)
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	
	// State should have recent timestamp
	if time.Since(state.Timestamp) > 2*time.Second {
		t.Errorf("State timestamp is too old: %v", state.Timestamp)
	}
	
	// State should have valid values
	if state.Position.Latitude < -90 || state.Position.Latitude > 90 {
		t.Errorf("Invalid latitude: %f", state.Position.Latitude)
	}
}

func TestSimulator_SubmitCommand_GoTo(t *testing.T) {
	simCfg, envCfg := createTestConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	
	sim, err := New(simCfg, envCfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond)
	
	cmd := models.NewCommand(models.CommandTypeGoTo)
	cmd.GoTo = &models.GoToCommand{
		Target: models.Position{
			Latitude:  32.1,
			Longitude: 34.1,
			Altitude:  1500.0,
		},
		Speed: ptr(100.0),
	}
	
	cmdCtx, cmdCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cmdCancel()
	
	err = sim.SubmitCommand(cmdCtx, cmd)
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	
	// Wait for command to be processed
	time.Sleep(300 * time.Millisecond)
	
	// Check that aircraft is moving
	getCtx, getCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer getCancel()
	
	state, _ := sim.GetState(getCtx)
	if state.Velocity.GroundSpeed <= 0 {
		t.Error("Aircraft not moving after goto command")
	}
}

func TestSimulator_SubmitCommand_Trajectory(t *testing.T) {
	simCfg, envCfg := createTestConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	
	sim, err := New(simCfg, envCfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond)
	
	cmd := models.NewCommand(models.CommandTypeTrajectory)
	cmd.Trajectory = &models.TrajectoryCommand{
		Waypoints: []models.Waypoint{
			{Position: models.Position{Latitude: 32.05, Longitude: 34.05, Altitude: 1200}, Speed: ptr(50.0)},
			{Position: models.Position{Latitude: 32.1, Longitude: 34.1, Altitude: 1500}, Speed: ptr(100.0)},
		},
		Loop: false,
	}
	
	cmdCtx, cmdCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cmdCancel()
	
	err = sim.SubmitCommand(cmdCtx, cmd)
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	
	time.Sleep(300 * time.Millisecond)
	
	getCtx, getCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer getCancel()
	
	state, _ := sim.GetState(getCtx)
	if state.Velocity.GroundSpeed <= 0 {
		t.Error("Aircraft not moving after trajectory command")
	}
}

func TestSimulator_GetPublisher(t *testing.T) {
	simCfg, envCfg := createTestConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	
	sim, err := New(simCfg, envCfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	
	pub := sim.GetPublisher()
	if pub == nil {
		t.Error("GetPublisher() returned nil")
	}
}

func TestSimulator_Run_Context_Cancellation(t *testing.T) {
	simCfg, envCfg := createTestConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	
	sim, err := New(simCfg, envCfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	
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

func ptr(f float64) *float64 {
	return &f
}
