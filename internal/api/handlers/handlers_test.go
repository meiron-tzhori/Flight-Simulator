package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

// Helper function to create a test simulator
func createTestSimulator(t *testing.T) *simulator.Simulator {
	t.Helper()
	
	cfg := &config.Config{
		Initial: config.InitialState{
			Latitude:  32.0,
			Longitude: 34.0,
			Altitude:  1000.0,
			Heading:   0.0,
		},
		Aircraft: config.AircraftConfig{
			MaxSpeed:         250.0,
			MaxClimbRate:     15.0,
			MaxDescentRate:   10.0,
			TurnRate:         3.0,
			CruiseSpeed:      100.0,
			CruiseAltitude:   1500.0,
			MinSpeed:         30.0,
			MaxAcceleration:  5.0,
		},
		Environment: config.EnvironmentConfig{
			WindDirection: 270.0,
			WindSpeed:     0.0,
		},
		Simulation: config.SimulationConfig{
			UpdateInterval:   100 * time.Millisecond,
			CommandQueueSize: 10,
		},
	}
	
	sim := simulator.NewSimulator(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	
	go sim.Run(ctx)
	time.Sleep(50 * time.Millisecond) // Let simulator start
	
	return sim
}

func TestHealthHandler(t *testing.T) {
	sim := createTestSimulator(t)
	handler := NewHandler(sim)
	
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	handler.Health(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Health() status = %d, want %d", w.Code, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Health() status = %v, want 'healthy'", response["status"])
	}
	
	if _, ok := response["timestamp"]; !ok {
		t.Error("Health() missing timestamp field")
	}
}

func TestGetStateHandler(t *testing.T) {
	sim := createTestSimulator(t)
	handler := NewHandler(sim)
	
	req := httptest.NewRequest(http.MethodGet, "/state", nil)
	w := httptest.NewRecorder()
	
	handler.GetState(w, req)
	
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
	
	if state.Position.Longitude != 34.0 {
		t.Errorf("GetState() longitude = %f, want 34.0", state.Position.Longitude)
	}
	
	if state.Position.Altitude != 1000.0 {
		t.Errorf("GetState() altitude = %f, want 1000.0", state.Position.Altitude)
	}
	
	// Check timestamp is recent
	if time.Since(state.Timestamp) > 2*time.Second {
		t.Errorf("GetState() timestamp too old: %v", state.Timestamp)
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
			payload: models.GoToCommand{
				Target: models.Position{
					Latitude:  32.1,
					Longitude: 34.1,
					Altitude:  1500.0,
				},
				Speed: ptr(100.0),
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Valid goto without speed",
			payload: models.GoToCommand{
				Target: models.Position{
					Latitude:  32.05,
					Longitude: 34.05,
					Altitude:  1200.0,
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Invalid latitude",
			payload: models.GoToCommand{
				Target: models.Position{
					Latitude:  95.0,
					Longitude: 34.1,
					Altitude:  1500.0,
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid speed",
			payload: models.GoToCommand{
				Target: models.Position{
					Latitude:  32.1,
					Longitude: 34.1,
					Altitude:  1500.0,
				},
				Speed: ptr(300.0),
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid JSON",
			payload:    "not json",
			wantStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := createTestSimulator(t)
			handler := NewHandler(sim)
			
			var body []byte
			var err error
			
			if str, ok := tt.payload.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("Failed to marshal payload: %v", err)
				}
			}
			
			req := httptest.NewRequest(http.MethodPost, "/command/goto", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			handler.GoToCommand(w, req)
			
			if w.Code != tt.wantStatus {
				t.Errorf("GoToCommand() status = %d, want %d. Body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestTrajectoryCommandHandler(t *testing.T) {
	tests := []struct {
		name       string
		payload    interface{}
		wantStatus int
	}{
		{
			name: "Valid trajectory",
			payload: models.TrajectoryCommand{
				Waypoints: []models.Waypoint{
					{
						Position: models.Position{Latitude: 32.0, Longitude: 34.0, Altitude: 1000},
						Speed:    ptr(50.0),
					},
					{
						Position: models.Position{Latitude: 32.1, Longitude: 34.1, Altitude: 1500},
						Speed:    ptr(100.0),
					},
				},
				Loop: false,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Valid looping trajectory",
			payload: models.TrajectoryCommand{
				Waypoints: []models.Waypoint{
					{
						Position: models.Position{Latitude: 32.0, Longitude: 34.0, Altitude: 1000},
					},
					{
						Position: models.Position{Latitude: 32.1, Longitude: 34.1, Altitude: 1000},
					},
				},
				Loop: true,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Empty waypoints",
			payload: models.TrajectoryCommand{
				Waypoints: []models.Waypoint{},
				Loop:      false,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid waypoint",
			payload: models.TrajectoryCommand{
				Waypoints: []models.Waypoint{
					{
						Position: models.Position{Latitude: 95.0, Longitude: 34.0, Altitude: 1000},
					},
				},
				Loop: false,
			},
			wantStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := createTestSimulator(t)
			handler := NewHandler(sim)
			
			body, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}
			
			req := httptest.NewRequest(http.MethodPost, "/command/trajectory", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			handler.TrajectoryCommand(w, req)
			
			if w.Code != tt.wantStatus {
				t.Errorf("TrajectoryCommand() status = %d, want %d. Body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestStopCommandHandler(t *testing.T) {
	sim := createTestSimulator(t)
	handler := NewHandler(sim)
	
	req := httptest.NewRequest(http.MethodPost, "/command/stop", nil)
	w := httptest.NewRecorder()
	
	handler.StopCommand(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("StopCommand() status = %d, want %d", w.Code, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if _, ok := response["message"]; !ok {
		t.Error("StopCommand() missing message field")
	}
}

func TestHoldCommandHandler(t *testing.T) {
	sim := createTestSimulator(t)
	handler := NewHandler(sim)
	
	req := httptest.NewRequest(http.MethodPost, "/command/hold", nil)
	w := httptest.NewRecorder()
	
	handler.HoldCommand(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("HoldCommand() status = %d, want %d", w.Code, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if _, ok := response["message"]; !ok {
		t.Error("HoldCommand() missing message field")
	}
}

func TestStreamHandler(t *testing.T) {
	sim := createTestSimulator(t)
	handler := NewHandler(sim)
	
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	w := httptest.NewRecorder()
	
	// Run stream handler in goroutine
	done := make(chan bool)
	go func() {
		handler.Stream(w, req)
		done <- true
	}()
	
	// Wait a bit for stream to start
	time.Sleep(200 * time.Millisecond)
	
	// Cancel context to stop streaming
	req.Context().Done()
	
	// Wait for handler to finish or timeout
	select {
	case <-done:
		// Handler finished
	case <-time.After(2 * time.Second):
		t.Error("Stream handler did not stop within timeout")
	}
	
	// Check headers
	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Stream() Content-Type = %s, want text/event-stream", ct)
	}
	
	if cc := w.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Stream() Cache-Control = %s, want no-cache", cc)
	}
	
	// Check body contains SSE data
	body := w.Body.String()
	if !strings.Contains(body, "data:") {
		t.Error("Stream() body does not contain SSE data events")
	}
}

func TestCommandSequence(t *testing.T) {
	// Test a sequence of commands
	sim := createTestSimulator(t)
	handler := NewHandler(sim)
	
	// 1. Get initial state
	req := httptest.NewRequest(http.MethodGet, "/state", nil)
	w := httptest.NewRecorder()
	handler.GetState(w, req)
	
	var initialState models.AircraftState
	if err := json.NewDecoder(w.Body).Decode(&initialState); err != nil {
		t.Fatalf("Failed to decode initial state: %v", err)
	}
	
	// 2. Send goto command
	gotoCmd := models.GoToCommand{
		Target: models.Position{
			Latitude:  32.1,
			Longitude: 34.1,
			Altitude:  1500.0,
		},
		Speed: ptr(100.0),
	}
	
	body, _ := json.Marshal(gotoCmd)
	req = httptest.NewRequest(http.MethodPost, "/command/goto", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.GoToCommand(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("GoToCommand failed: %s", w.Body.String())
	}
	
	// 3. Wait for simulation to update
	time.Sleep(300 * time.Millisecond)
	
	// 4. Get state again and verify aircraft moved
	req = httptest.NewRequest(http.MethodGet, "/state", nil)
	w = httptest.NewRecorder()
	handler.GetState(w, req)
	
	var newState models.AircraftState
	if err := json.NewDecoder(w.Body).Decode(&newState); err != nil {
		t.Fatalf("Failed to decode new state: %v", err)
	}
	
	// Aircraft should have moved (position or velocity changed)
	positionChanged := newState.Position.Latitude != initialState.Position.Latitude ||
		newState.Position.Longitude != initialState.Position.Longitude ||
		newState.Position.Altitude != initialState.Position.Altitude
	
	velocityNonZero := newState.Velocity.GroundSpeed > 0 || newState.Velocity.VerticalSpeed != 0
	
	if !positionChanged && !velocityNonZero {
		t.Error("Aircraft did not move after goto command")
	}
	
	// 5. Send stop command
	req = httptest.NewRequest(http.MethodPost, "/command/stop", nil)
	w = httptest.NewRecorder()
	handler.StopCommand(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("StopCommand failed: %s", w.Body.String())
	}
}

// Helper function
func ptr(f float64) *float64 {
	return &f
}
