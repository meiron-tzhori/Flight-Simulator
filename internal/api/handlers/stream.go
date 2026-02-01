package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
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
func (h *StreamHandler) Stream(c *gin.Context) {
	// TODO: Implement SSE streaming
	// 1. Set SSE headers
	// 2. Subscribe to state publisher
	// 3. Stream state updates
	// 4. Cleanup on disconnect

	c.String(200, "SSE streaming not implemented yet\n")
}
