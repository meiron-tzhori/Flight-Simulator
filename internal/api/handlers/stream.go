package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

// StreamHandler handles SSE streaming requests.
type StreamHandler struct {
	simulator *simulator.Simulator
	logger    *slog.Logger
}

// NewStreamHandler creates a new stream handler.
func NewStreamHandler(sim *simulator.Simulator, logger *slog.Logger) *StreamHandler {
	return &StreamHandler{
		simulator: sim,
		logger:    logger,
	}
}

// Stream handles GET /stream
// Streams aircraft state updates via Server-Sent Events (SSE).
func (h *StreamHandler) Stream(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Check if response writer supports flushing
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		h.logger.Error("Streaming not supported - response writer doesn't support flushing")
		c.String(http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// Generate unique subscriber ID
	subID := uuid.New().String()

	// Subscribe to state updates
	publisher := h.simulator.GetPublisher()
	stateChan := publisher.Subscribe(subID)
	defer publisher.Unsubscribe(subID)

	h.logger.Info("SSE client connected", "subscriber_id", subID, "remote_addr", c.ClientIP())
	defer h.logger.Info("SSE client disconnected", "subscriber_id", subID)

	// Send initial connection event
	fmt.Fprintf(c.Writer, "event: connected\n")
	fmt.Fprintf(c.Writer, "data: {\"subscriber_id\":\"%s\"}\n\n", subID)
	flusher.Flush()

	// Throttle updates to 10 Hz (every 100ms) to avoid overwhelming clients
	throttle := time.NewTicker(100 * time.Millisecond)
	defer throttle.Stop()

	// Keep track of latest state
	var latestState *models.AircraftState

	// Heartbeat to detect client disconnections
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case state, ok := <-stateChan:
			if !ok {
				// Channel closed (simulator shutdown)
				h.logger.Info("State channel closed", "subscriber_id", subID)
				return
			}
			// Cache latest state (will be sent on next throttle tick)
			latestState = &state

		case <-throttle.C:
			if latestState != nil {
				// Marshal state to JSON
				data, err := json.Marshal(latestState)
				if err != nil {
					h.logger.Error("Failed to marshal state", "error", err, "subscriber_id", subID)
					continue
				}

				// Send SSE event in standard format:
				// event: state
				// data: {json}
				// (blank line)
				fmt.Fprintf(c.Writer, "event: state\n")
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)

				// Flush to send immediately
				flusher.Flush()

				// Check if client disconnected
				if c.Writer.Written() < 0 {
					h.logger.Info("Client disconnected (write error)", "subscriber_id", subID)
					return
				}
			}

		case <-heartbeat.C:
			// Send heartbeat to keep connection alive
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			flusher.Flush()

		case <-c.Request.Context().Done():
			// Client disconnected or request cancelled
			h.logger.Info("Client request context done", "subscriber_id", subID)
			return
		}
	}
}
