package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

// DependencyResolver resolves project dependencies from the global hub and clones them.
type DependencyResolver struct {
	Hub *config.HubConfig
}

// NewDependencyResolver loads the global hub configuration.
func NewDependencyResolver() (*DependencyResolver, error) {
	hub, err := config.LoadGlobalHub()
	if err != nil {
		return nil, err
	}
	return &DependencyResolver{Hub: hub}, nil
}

// ResolveAndClone resolves and clones each required dependency.
func (r *DependencyResolver) ResolveAndClone(projectRoot string, requiredAliases []string) error {
	parentDir := filepath.Dir(projectRoot)

	for _, alias := range requiredAliases {
		targetPath := filepath.Join(parentDir, alias)

		if _, err := os.Stat(targetPath); err == nil {
			ui.Infof("Dependency '%s' already exists locally", alias)
			continue
		}

		gitURL, err := r.Hub.ResolveAlias(alias)
		if err != nil {
			return fmt.Errorf("failed to resolve dependency '%s': %w", alias, err)
		}

		ui.Taskf("Cloning dependency '%s' from %s", alias, gitURL)

		cmd := exec.Command("git", "clone", gitURL, targetPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			ui.Error("FAILED")
			return fmt.Errorf("failed to clone '%s': %w", alias, err)
		}

		ui.Success("DONE")
	}

	return nil
}
