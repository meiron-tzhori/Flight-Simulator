package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration.
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Simulation SimulationConfig `yaml:"simulation"`
	Environment EnvironmentConfig `yaml:"environment"`
	Logging    LoggingConfig    `yaml:"logging"`
	Metrics    MetricsConfig    `yaml:"metrics"`
	Streaming  StreamingConfig  `yaml:"streaming"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// SimulationConfig contains simulation engine settings.
type SimulationConfig struct {
	TickRateHz         float64          `yaml:"tick_rate_hz"`
	CommandQueueSize   int              `yaml:"command_queue_size"`
	InitialPosition    PositionConfig   `yaml:"initial_position"`
	InitialVelocity    VelocityConfig   `yaml:"initial_velocity"`
	InitialHeading     float64          `yaml:"initial_heading"`
	DefaultSpeed       float64          `yaml:"default_speed"`
	MaxSpeed           float64          `yaml:"max_speed"`
	MaxClimbRate       float64          `yaml:"max_climb_rate"`
	MaxDescentRate     float64          `yaml:"max_descent_rate"`
	PositionTolerance  float64          `yaml:"position_tolerance"`
	HeadingChangeRate  float64          `yaml:"heading_change_rate"`
	SpeedChangeRate    float64          `yaml:"speed_change_rate"`
}

// PositionConfig represents a configured position.
type PositionConfig struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
	Altitude  float64 `yaml:"altitude"`
}

// VelocityConfig represents a configured velocity.
type VelocityConfig struct {
	GroundSpeed   float64 `yaml:"ground_speed"`
	VerticalSpeed float64 `yaml:"vertical_speed"`
}

// EnvironmentConfig contains environment settings.
type EnvironmentConfig struct {
	Enabled bool        `yaml:"enabled"`
	Wind    WindConfig  `yaml:"wind"`
	Humidity HumidityConfig `yaml:"humidity"`
	Terrain TerrainConfig `yaml:"terrain"`
}

// WindConfig contains wind settings.
type WindConfig struct {
	Enabled   bool    `yaml:"enabled"`
	Direction float64 `yaml:"direction"`
	Speed     float64 `yaml:"speed"`
}

// HumidityConfig contains humidity settings.
type HumidityConfig struct {
	Enabled bool    `yaml:"enabled"`
	Value   float64 `yaml:"value"`
}

// TerrainConfig contains terrain settings.
type TerrainConfig struct {
	Enabled      bool    `yaml:"enabled"`
	SafetyMargin float64 `yaml:"safety_margin"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level         string `yaml:"level"`
	Format        string `yaml:"format"`
	IncludeCaller bool   `yaml:"include_caller"`
}

// MetricsConfig contains metrics settings.
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// StreamingConfig contains SSE streaming settings.
type StreamingConfig struct {
	Enabled      bool `yaml:"enabled"`
	UpdateRateHz int  `yaml:"update_rate_hz"`
	BufferSize   int  `yaml:"buffer_size"`
	MaxClients   int  `yaml:"max_clients"`
}

// Load loads configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// TODO: Apply environment variable overrides
	// TODO: Validate configuration

	return &cfg, nil
}
