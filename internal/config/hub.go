package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type HubConfig struct {
	Registries []string          `yaml:"registries,omitempty"`
	Projects   map[string]string `yaml:"projects,omitempty"`
}

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
