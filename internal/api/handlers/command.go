package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

// CommandHandler handles command requests.
type CommandHandler struct {
	simulator *simulator.Simulator
	logger    *slog.Logger
}

// NewCommandHandler creates a new command handler.
func NewCommandHandler(sim *simulator.Simulator, logger *slog.Logger) *CommandHandler {
	return &CommandHandler{
		simulator: sim,
		logger:    logger,
	}
}

// GoTo handles POST /command/goto
func (h *CommandHandler) GoTo(c *gin.Context) {
	// TODO: Implement go-to command handler
	// 1. Parse and validate request
	// 2. Create command
	// 3. Submit to simulator
	// 4. Return response

	c.JSON(http.StatusOK, models.CommandResponse{
		Status:  "accepted",
		Message: "Go-to command accepted (not implemented yet)",
	})
}

// Trajectory handles POST /command/trajectory
func (h *CommandHandler) Trajectory(c *gin.Context) {
	// TODO: Implement trajectory command handler
	c.JSON(http.StatusOK, models.CommandResponse{
		Status:  "accepted",
		Message: "Trajectory command accepted (not implemented yet)",
	})
}

// Stop handles POST /command/stop
func (h *CommandHandler) Stop(c *gin.Context) {
	// TODO: Implement stop command handler
	c.JSON(http.StatusOK, models.CommandResponse{
		Status:  "accepted",
		Message: "Stop command accepted (not implemented yet)",
	})
}

// Hold handles POST /command/hold
func (h *CommandHandler) Hold(c *gin.Context) {
	// TODO: Implement hold command handler
	c.JSON(http.StatusOK, models.CommandResponse{
		Status:  "accepted",
		Message: "Hold command accepted (not implemented yet)",
	})
}
