package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// HubConfig stores global hub configuration for projects.
type HubConfig struct {
	Projects map[string]string `yaml:"projects,omitempty"`
}

// LoadGlobalHub loads the global hub configuration from the user's home directory.
func LoadGlobalHub() (*HubConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".derrick", "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &HubConfig{
			Projects: make(map[string]string),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global hub config: %w", err)
	}

	var hub HubConfig
	if err := yaml.Unmarshal(data, &hub); err != nil {
		return nil, fmt.Errorf("failed to parse global hub config at %s: %w", configPath, err)
	}

	if hub.Projects == nil {
		hub.Projects = make(map[string]string)
	}

	return &hub, nil
}

func (h *HubConfig) ResolveAlias(alias string) (string, error) {
	if url, exists := h.Projects[alias]; exists {
		return url, nil
	}

	return "", fmt.Errorf("project alias '%s' not found in local Derrick hub", alias)
}

// Save writes the hub configuration back to the user's home directory.
func (h *HubConfig) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".derrick")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, err := yaml.Marshal(h)
	if err != nil {
		return fmt.Errorf("failed to marshal hub config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write hub config to %s: %w", configPath, err)
	}

	return nil
}
