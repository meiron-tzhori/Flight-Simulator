package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

// StateHandler handles state query requests.
type StateHandler struct {
	simulator *simulator.Simulator
	logger    *slog.Logger
}

// NewStateHandler creates a new state handler.
func NewStateHandler(sim *simulator.Simulator, logger *slog.Logger) *StateHandler {
	return &StateHandler{
		simulator: sim,
		logger:    logger,
	}
}

// GetState handles GET /state
func (h *StateHandler) GetState(c *gin.Context) {
	state, err := h.simulator.GetState(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get state", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve aircraft state",
		})
		return
	}

	c.JSON(http.StatusOK, state)
}
