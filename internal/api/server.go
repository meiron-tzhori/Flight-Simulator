package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meiron-tzhori/Flight-Simulator/internal/api/handlers"
	"github.com/meiron-tzhori/Flight-Simulator/internal/api/middleware"
	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

// Server represents the HTTP API server.
type Server struct {
	httpServer *http.Server
	simulator  *simulator.Simulator
	logger     *slog.Logger
}

// NewServer creates a new API server.
func NewServer(cfg config.ServerConfig, sim *simulator.Simulator, logger *slog.Logger) *Server {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS())

	// Create handlers
	healthHandler := handlers.NewHealthHandler(sim, logger)
	commandHandler := handlers.NewCommandHandler(sim, logger)
	stateHandler := handlers.NewStateHandler(sim, logger)
	streamHandler := handlers.NewStreamHandler(sim, logger)

	// Register routes
	router.GET("/health", healthHandler.Health)
	router.GET("/state", stateHandler.GetState)
	router.GET("/stream", streamHandler.Stream)
	router.POST("/command/goto", commandHandler.GoTo)
	router.POST("/command/trajectory", commandHandler.Trajectory)
	router.POST("/command/stop", commandHandler.Stop)
	router.POST("/command/hold", commandHandler.Hold)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &Server{
		httpServer: httpServer,
		simulator:  sim,
		logger:     logger,
	}
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting HTTP server", "addr", s.httpServer.Addr)

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}
