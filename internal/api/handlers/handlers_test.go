package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

// Helper function to create a test simulator
func createTestSimulator(t *testing.T) *simulator.Simulator {
	t.Helper()
	
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
	
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	
	sim, err := simulator.New(simCfg, envCfg, logger)
	if err != nil {
		t.Fatalf("Failed to create simulator: %v", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	
	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond) // Let simulator start
	
	return sim
}

func setupRouter(sim *simulator.Simulator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	
	// Setup handlers with correct constructors
	cmdHandler := NewCommandHandler(sim, logger, 250.0)
	stateHandler := NewStateHandler(sim, logger)
	healthHandler := NewHealthHandler(sim, logger, 10.0) // tickRate = 10 Hz
	streamHandler := NewStreamHandler(sim, logger)
	
	router.GET("/health", healthHandler.Health)
	router.GET("/state", stateHandler.GetState)
	router.POST("/command/goto", cmdHandler.GoTo)
	router.POST("/command/trajectory", cmdHandler.Trajectory)
	router.POST("/command/stop", cmdHandler.Stop)
	router.POST("/command/hold", cmdHandler.Hold)
	router.GET("/stream", streamHandler.Stream)
	
	return router
}

func TestHealthHandler(t *testing.T) {
	sim := createTestSimulator(t)
	router := setupRouter(sim)
	
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Health() status = %d, want %d", w.Code, http.StatusOK)
	}
	
	var response models.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response.Status != "healthy" {
		t.Errorf("Health() status = %v, want 'healthy'", response.Status)
	}
}

func TestGetStateHandler(t *testing.T) {
	sim := createTestSimulator(t)
	router := setupRouter(sim)
	
	req := httptest.NewRequest(http.MethodGet, "/state", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("GetState() status = %d, want %d", w.Code, http.StatusOK)
	}
	
	var state models.AircraftState
	if err := json.NewDecoder(w.Body).Decode(&state); err != nil {
		t.Fatalf("Failed to decode state: %v", err)
	}
	
	// Check initial position
	if state.Position.Latitude != 32.0 {
		t.Errorf("GetState() latitude = %f, want 32.0", state.Position.Latitude)
	}
}

func TestGoToCommandHandler(t *testing.T) {
	tests := []struct {
		name       string
		payload    interface{}
		wantStatus int
	}{
		{
			name: "Valid goto command",
			payload: GoToRequest{
				Lat:   32.1,
				Lon:   34.1,
				Alt:   1500.0,
				Speed: ptr(100.0),
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Invalid latitude",
			payload: GoToRequest{
				Lat: 95.0,
				Lon: 34.1,
				Alt: 1500.0,
			},
			wantStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := createTestSimulator(t)
			router := setupRouter(sim)
			
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/command/goto", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			if w.Code != tt.wantStatus {
				t.Errorf("GoTo() status = %d, want %d. Body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestTrajectoryCommandHandler(t *testing.T) {
	tests := []struct {
		name       string
		payload    TrajectoryRequest
		wantStatus int
	}{
		{
			name: "Valid trajectory",
			payload: TrajectoryRequest{
				Waypoints: []WaypointRequest{
					{Lat: 32.0, Lon: 34.0, Alt: 1000, Speed: ptr(50.0)},
					{Lat: 32.1, Lon: 34.1, Alt: 1500, Speed: ptr(100.0)},
				},
				Loop: false,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Empty waypoints",
			payload: TrajectoryRequest{
				Waypoints: []WaypointRequest{},
				Loop:      false,
			},
			wantStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := createTestSimulator(t)
			router := setupRouter(sim)
			
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/command/trajectory", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			if w.Code != tt.wantStatus {
				t.Errorf("Trajectory() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestStopCommandHandler(t *testing.T) {
	sim := createTestSimulator(t)
	router := setupRouter(sim)
	
	req := httptest.NewRequest(http.MethodPost, "/command/stop", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Stop() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHoldCommandHandler(t *testing.T) {
	sim := createTestSimulator(t)
	router := setupRouter(sim)
	
	req := httptest.NewRequest(http.MethodPost, "/command/hold", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Hold() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCommandSequence(t *testing.T) {
	sim := createTestSimulator(t)
	router := setupRouter(sim)
	
	// 1. Get initial state
	req := httptest.NewRequest(http.MethodGet, "/state", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	var initialState models.AircraftState
	if err := json.NewDecoder(w.Body).Decode(&initialState); err != nil {
		t.Fatalf("Failed to decode initial state: %v", err)
	}
	
	// 2. Send goto command
	gotoReq := GoToRequest{
		Lat:   32.1,
		Lon:   34.1,
		Alt:   1500.0,
		Speed: ptr(100.0),
	}
	
	body, _ := json.Marshal(gotoReq)
	req = httptest.NewRequest(http.MethodPost, "/command/goto", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("GoTo failed: %s", w.Body.String())
	}
	
	// 3. Wait for simulation to update
	time.Sleep(300 * time.Millisecond)
	
	// 4. Get state again - aircraft should be moving
	req = httptest.NewRequest(http.MethodGet, "/state", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	var newState models.AircraftState
	if err := json.NewDecoder(w.Body).Decode(&newState); err != nil {
		t.Fatalf("Failed to decode new state: %v", err)
	}
	
	// Aircraft should have moved or have velocity
	if newState.Velocity.GroundSpeed == 0 && newState.Position == initialState.Position {
		t.Error("Aircraft did not respond to goto command")
	}
}

func ptr(f float64) *float64 {
	return &f
}
