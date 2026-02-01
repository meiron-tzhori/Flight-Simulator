package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meiron-tzhori/Flight-Simulator/internal/api/validation"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
	"github.com/meiron-tzhori/Flight-Simulator/pkg/geo"
)

// CommandHandler handles command requests.
type CommandHandler struct {
	simulator *simulator.Simulator
	logger    *slog.Logger
	maxSpeed  float64
}

// NewCommandHandler creates a new command handler.
func NewCommandHandler(sim *simulator.Simulator, logger *slog.Logger, maxSpeed float64) *CommandHandler {
	return &CommandHandler{
		simulator: sim,
		logger:    logger,
		maxSpeed:  maxSpeed,
	}
}

// GoToRequest represents the request body for go-to command.
type GoToRequest struct {
	Lat   float64  `json:"lat" binding:"required"`
	Lon   float64  `json:"lon" binding:"required"`
	Alt   float64  `json:"alt" binding:"required"`
	Speed *float64 `json:"speed,omitempty"`
}

// GoTo handles POST /command/goto
func (h *CommandHandler) GoTo(c *gin.Context) {
	var req GoToRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Create command
	cmd := models.NewCommand(models.CommandTypeGoTo)
	cmd.GoTo = &models.GoToCommand{
		Target: models.Position{
			Latitude:  req.Lat,
			Longitude: req.Lon,
			Altitude:  req.Alt,
		},
		Speed: req.Speed,
	}

	// Validate
	if err := validation.ValidateGoToCommand(cmd.GoTo, h.maxSpeed); err != nil {
		h.logger.Warn("Validation failed", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    getErrorCode(err),
				Message: err.Error(),
			},
		})
		return
	}

	// Submit to simulator
	if err := h.simulator.SubmitCommand(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to submit command", "error", err)
		if errors.Is(err, models.ErrCommandQueueFull) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "QUEUE_FULL",
					Message: "Command queue is full, please retry",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to submit command",
				},
			})
		}
		return
	}

	// Get current state for ETA calculation
	state, err := h.simulator.GetState(c.Request.Context())
	if err != nil {
		h.logger.Warn("Failed to get state for ETA calculation", "error", err)
	}

	// Calculate ETA
	var etaSeconds float64
	if err == nil {
		distance := geo.Haversine(
			state.Position.Latitude,
			state.Position.Longitude,
			cmd.GoTo.Target.Latitude,
			cmd.GoTo.Target.Longitude,
		)
		speed := state.Velocity.GroundSpeed
		if speed == 0 && cmd.GoTo.Speed != nil {
			speed = *cmd.GoTo.Speed
		}
		if speed > 0 {
			etaSeconds = distance / speed
		}
	}

	// Success
	c.JSON(http.StatusOK, models.CommandResponse{
		Status:     "accepted",
		CommandID:  cmd.ID,
		Message:    "Go-to command accepted",
		Target:     &cmd.GoTo.Target,
		ETASeconds: etaSeconds,
	})
}

// TrajectoryRequest represents the request body for trajectory command.
type TrajectoryRequest struct {
	Waypoints []WaypointRequest `json:"waypoints" binding:"required,min=1"`
	Loop      bool              `json:"loop"`
}

type WaypointRequest struct {
	Lat   float64  `json:"lat" binding:"required"`
	Lon   float64  `json:"lon" binding:"required"`
	Alt   float64  `json:"alt" binding:"required"`
	Speed *float64 `json:"speed,omitempty"`
}

// Trajectory handles POST /command/trajectory
func (h *CommandHandler) Trajectory(c *gin.Context) {
	var req TrajectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Create command
	cmd := models.NewCommand(models.CommandTypeTrajectory)
	waypoints := make([]models.Waypoint, len(req.Waypoints))
	for i, wp := range req.Waypoints {
		waypoints[i] = models.Waypoint{
			Position: models.Position{
				Latitude:  wp.Lat,
				Longitude: wp.Lon,
				Altitude:  wp.Alt,
			},
			Speed: wp.Speed,
		}
	}
	cmd.Trajectory = &models.TrajectoryCommand{
		Waypoints: waypoints,
		Loop:      req.Loop,
	}

	// Validate
	if err := validation.ValidateTrajectoryCommand(cmd.Trajectory, h.maxSpeed); err != nil {
		h.logger.Warn("Validation failed", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    getErrorCode(err),
				Message: err.Error(),
			},
		})
		return
	}

	// Submit to simulator
	if err := h.simulator.SubmitCommand(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to submit command", "error", err)
		if errors.Is(err, models.ErrCommandQueueFull) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "QUEUE_FULL",
					Message: "Command queue is full, please retry",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to submit command",
				},
			})
		}
		return
	}

	// Success
	c.JSON(http.StatusOK, models.CommandResponse{
		Status:        "accepted",
		CommandID:     cmd.ID,
		Message:       "Trajectory command accepted",
		WaypointCount: len(waypoints),
	})
}

// Stop handles POST /command/stop
func (h *CommandHandler) Stop(c *gin.Context) {
	cmd := models.NewCommand(models.CommandTypeStop)

	if err := h.simulator.SubmitCommand(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to submit command", "error", err)
		if errors.Is(err, models.ErrCommandQueueFull) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "QUEUE_FULL",
					Message: "Command queue is full, please retry",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to submit stop command",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.CommandResponse{
		Status:    "accepted",
		CommandID: cmd.ID,
		Message:   "Stop command accepted",
	})
}

// Hold handles POST /command/hold
func (h *CommandHandler) Hold(c *gin.Context) {
	cmd := models.NewCommand(models.CommandTypeHold)

	if err := h.simulator.SubmitCommand(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to submit command", "error", err)
		if errors.Is(err, models.ErrCommandQueueFull) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "QUEUE_FULL",
					Message: "Command queue is full, please retry",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to submit hold command",
				},
			})
		}
		return
	}

	// Get current position for response
	state, err := h.simulator.GetState(c.Request.Context())
	if err != nil {
		h.logger.Warn("Failed to get state for hold response", "error", err)
	}

	response := models.CommandResponse{
		Status:    "accepted",
		CommandID: cmd.ID,
		Message:   "Hold command accepted",
	}

	if err == nil {
		response.HoldPosition = &state.Position
		response.OrbitRadiusM = 0 // Simple hold, no orbit
	}

	c.JSON(http.StatusOK, response)
}

// getErrorCode extracts error code from error.
func getErrorCode(err error) string {
	switch {
	case errors.Is(err, models.ErrInvalidLatitude):
		return "INVALID_LATITUDE"
	case errors.Is(err, models.ErrInvalidLongitude):
		return "INVALID_LONGITUDE"
	case errors.Is(err, models.ErrInvalidAltitude):
		return "INVALID_ALTITUDE"
	case errors.Is(err, models.ErrInvalidSpeed):
		return "INVALID_SPEED"
	case errors.Is(err, models.ErrEmptyWaypoints):
		return "EMPTY_WAYPOINTS"
	case errors.Is(err, models.ErrSpeedExceedsMax):
		return "SPEED_EXCEEDS_MAX"
	default:
		return "VALIDATION_ERROR"
	}
}
