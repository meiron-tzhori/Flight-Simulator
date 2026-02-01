package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	simulator *simulator.Simulator
	logger    *slog.Logger
	startTime time.Time
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(sim *simulator.Simulator, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		simulator: sim,
		logger:    logger,
		startTime: time.Now(),
	}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	// TODO: Add actual health checks
	response := models.HealthResponse{
		Status:            "healthy",
		SimulationRunning: true,
		TickRateHz:        30.0, // TODO: Get from config
		UptimeSeconds:     time.Since(h.startTime).Seconds(),
		Timestamp:         time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
