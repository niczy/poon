package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config represents the poon workspace configuration
type Config struct {
	WorkspaceName string   `json:"workspaceName"`
	GitServerURL  string   `json:"gitServerUrl"`
	GrpcServerURL string   `json:"grpcServerUrl"`
	TrackedPaths  []string `json:"trackedPaths"`
	CreatedAt     string   `json:"createdAt"`
}

// TrackedPath represents a tracked path with metadata
type TrackedPath struct {
	Path         string `json:"path"`
	LastSyncHash string `json:"lastSyncHash"`
	AddedAt      string `json:"addedAt"`
}

// LoadConfig loads the poon configuration from .poon/config.json
func LoadConfig() (*Config, error) {
	configPath := ".poon/config.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no poon workspace found (run 'poon start' first)")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

// SaveConfig saves the poon configuration to .poon/config.json
func SaveConfig(config *Config) error {
	if err := os.MkdirAll(".poon", 0755); err != nil {
		return fmt.Errorf("failed to create .poon directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	configPath := ".poon/config.json"
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	return nil
}

// CreateConfig creates a new poon configuration
func CreateConfig(workspaceName, gitServerURL, grpcServerURL string, trackedPaths []string) *Config {
	return &Config{
		WorkspaceName: workspaceName,
		GitServerURL:  gitServerURL,
		GrpcServerURL: grpcServerURL,
		TrackedPaths:  trackedPaths,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
}
