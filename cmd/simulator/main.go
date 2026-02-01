package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/meiron-tzhori/Flight-Simulator/internal/api"
	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/observability"
	"github.com/meiron-tzhori/Flight-Simulator/internal/simulator"
)

var (
	version = "dev" // Set via ldflags during build
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Flight Simulator version %s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := observability.NewLogger(cfg.Logging)
	logger.Info("Starting Flight Simulator",
		"version", version,
		"config", *configPath,
	)

	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize components
	sim, err := simulator.New(cfg.Simulation, logger)
	if err != nil {
		logger.Error("Failed to create simulator", "error", err)
		os.Exit(1)
	}

	server := api.NewServer(cfg.Server, sim, logger)

	// Start components
	var wg sync.WaitGroup

	// Start simulator
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := sim.Run(ctx); err != nil {
			logger.Error("Simulator error", "error", err)
		}
	}()

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(ctx); err != nil {
			logger.Error("Server error", "error", err)
		}
	}()

	logger.Info("Flight Simulator is running",
		"http_port", cfg.Server.Port,
		"tick_rate_hz", cfg.Simulation.TickRateHz,
	)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutdown signal received, initiating graceful shutdown...")

	// Cancel context to signal shutdown
	cancel()

	// Wait for graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Graceful shutdown complete")
	case <-time.After(30 * time.Second):
		logger.Error("Shutdown timeout exceeded, forcing exit")
	}
}
