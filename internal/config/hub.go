package config

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// HubProject represents a project registered in the Hub.
type HubProject struct {
	URL  string `yaml:"url"`
	Path string `yaml:"path,omitempty"`
}

// UnmarshalYAML allows a HubProject to be defined as either a plain string (URL)
// or a full struct.
func (p *HubProject) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		p.URL = value.Value
		return nil
	}
	var tmp struct {
		URL  string `yaml:"url"`
		Path string `yaml:"path"`
	}
	if err := value.Decode(&tmp); err != nil {
		return err
	}
	p.URL = tmp.URL
	p.Path = tmp.Path
	return nil
}

// HubConfig stores global hub configuration for projects.
type HubConfig struct {
	Workspace string                `yaml:"workspace,omitempty"`
	Projects  map[string]HubProject `yaml:"projects,omitempty"`
	Remotes   []string              `yaml:"remotes,omitempty"`
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
			Projects: make(map[string]HubProject),
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
		hub.Projects = make(map[string]HubProject)
	}

	if hub.Workspace == "" {
		hub.Workspace = filepath.Join(homeDir, "derrick-projects")
	}

	return &hub, nil
}

func (h *HubConfig) ResolveAlias(alias string) (HubProject, error) {
	if proj, exists := h.Projects[alias]; exists {
		return proj, nil
	}

	// Try remotes
	for _, remote := range h.Remotes {
		client := http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(remote)
		if err != nil {
			continue // Skip unreachable remotes
		}

		if resp.StatusCode == 200 {
			var remoteHub HubConfig
			if err := yaml.NewDecoder(resp.Body).Decode(&remoteHub); err == nil {
				if proj, exists := remoteHub.Projects[alias]; exists {
					resp.Body.Close()
					return proj, nil
				}
			}
		}
		resp.Body.Close()
	}

	return HubProject{}, fmt.Errorf("project alias '%s' not found in local or remote Derrick hubs", alias)
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
