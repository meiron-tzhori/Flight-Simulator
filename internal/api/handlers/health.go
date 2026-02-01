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
	tickRate  float64
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(sim *simulator.Simulator, logger *slog.Logger, tickRate float64) *HealthHandler {
	return &HealthHandler{
		simulator: sim,
		logger:    logger,
		startTime: time.Now(),
		tickRate:  tickRate,
	}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	response := models.HealthResponse{
		Status:            "healthy",
		SimulationRunning: true,
		TickRateHz:        h.tickRate,
		UptimeSeconds:     time.Since(h.startTime).Seconds(),
		Timestamp:         time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
