package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ComposeMap represents the structure of a docker-compose.yml for parsing.
type ComposeMap struct {
	Services map[string]interface{} `yaml:"services"`
}

// OverrideMap represents the docker-compose.override.yml structure.
type OverrideMap struct {
	Services map[string]ServiceOverride `yaml:"services"`
}

// ServiceOverride defines service-specific overrides for docker-compose.
type ServiceOverride struct {
	ExtraHosts []string          `yaml:"extra_hosts,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty"`
}

// DerrickManagedLabel is applied to every derrick-managed docker resource so
// `derrick clean` can scope prune operations and never touch unrelated assets.
const DerrickManagedLabel = "com.derrick.managed=true"

// GenerateNetworkOverride creates a docker-compose.override.yml that injects
// host.docker.internal into every service's extra_hosts. The project's default
// network is left to Docker Compose so it remains scoped to this project and
// never conflicts with other Derrick projects.
func GenerateNetworkOverride(composeFile string, outDir string) (string, error) {
	data, err := os.ReadFile(composeFile)
	if err != nil {
		return "", err
	}

	var base ComposeMap
	if err := yaml.Unmarshal(data, &base); err != nil {
		return "", fmt.Errorf("failed to parse %s: %v", composeFile, err)
	}

	override := OverrideMap{
		Services: make(map[string]ServiceOverride),
	}

	for svcName := range base.Services {
		override.Services[svcName] = ServiceOverride{
			ExtraHosts: []string{"host.docker.internal:host-gateway"},
			Labels:     map[string]string{"com.derrick.managed": "true"},
		}
	}

	overrideData, err := yaml.Marshal(&override)
	if err != nil {
		return "", err
	}

	if outDir == "" {
		outDir = ".derrick"
	}

	err = os.MkdirAll(outDir, 0o755)
	if err != nil {
		return "", err
	}

	overridePath := filepath.Join(outDir, "docker-compose.override.yml")
	err = os.WriteFile(overridePath, overrideData, 0o644)
	if err != nil {
		return "", err
	}

	return overridePath, nil
}

// IsDockerInstalled checks if the docker binary is available in PATH.
func IsDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// FirstService returns the name of the first service defined in a compose file,
// preserving YAML declaration order. Used as the default exec target for `derrick shell`.
func FirstService(composeFile string) (string, error) {
	data, err := os.ReadFile(composeFile)
	if err != nil {
		return "", err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return "", fmt.Errorf("failed to parse %s: %w", composeFile, err)
	}
	if len(doc.Content) == 0 {
		return "", fmt.Errorf("empty compose file: %s", composeFile)
	}

	// Walk the root mapping to find the "services" key.
	root := doc.Content[0]
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "services" {
			svcMap := root.Content[i+1]
			if len(svcMap.Content) >= 2 {
				return svcMap.Content[0].Value, nil
			}
		}
	}
	return "", fmt.Errorf("no services found in %s", composeFile)
}

