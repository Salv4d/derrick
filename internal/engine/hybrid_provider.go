package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Salv4d/derrick/internal/config"
)

// providerLeg narrows the Provider interface to the subset HybridProvider
// composes. Exists so the hybrid logic is testable without a docker daemon
// or a nix installation.
type providerLeg interface {
	Name() string
	IsAvailable() error
	Start(cfg *config.ProjectConfig, flags Flags) error
	Stop(cfg *config.ProjectConfig) error
	Shell(cfg *config.ProjectConfig, args []string) error
	Status(cfg *config.ProjectConfig) (EnvironmentStatus, error)
}

// HybridProvider runs containers via Docker Compose while exposing language
// toolchains through a Nix dev shell.
//
// Contract:
//   - Start:  bring up containers, then materialize the Nix flake + .envrc.
//   - Stop:   tear down containers. Nix dev shells are process-scoped and
//     need no explicit stop.
//   - Shell:  drop into the Nix dev shell — this is where host-visible
//     language tools (go, node, …) live, not inside a container.
//   - Status: aggregate both legs and report them side-by-side.
type HybridProvider struct {
	docker providerLeg
	nix    providerLeg
}

// NewHybridProvider assembles the default hybrid implementation composing
// the concrete DockerProvider and NixProvider.
func NewHybridProvider() *HybridProvider {
	return &HybridProvider{
		docker: &DockerProvider{},
		nix:    &NixProvider{},
	}
}

func (h *HybridProvider) Name() string { return "hybrid" }

// IsAvailable requires both toolchains. Errors are joined so the user sees
// every missing dependency at once instead of discovering them one at a time.
func (h *HybridProvider) IsAvailable() error {
	var errs []error
	if err := h.docker.IsAvailable(); err != nil {
		errs = append(errs, fmt.Errorf("docker: %w", err))
	}
	if err := h.nix.IsAvailable(); err != nil {
		errs = append(errs, fmt.Errorf("nix: %w", err))
	}
	return errors.Join(errs...)
}

// Start brings up containers first, then materializes the flake. If docker
// fails we skip nix — a dev shell is useless when the backing services are
// down.
func (h *HybridProvider) Start(cfg *config.ProjectConfig, flags Flags) error {
	if err := h.docker.Start(cfg, flags); err != nil {
		return err
	}
	return h.nix.Start(cfg, flags)
}

// Stop tears down the docker leg. Nix is intentionally a no-op.
func (h *HybridProvider) Stop(cfg *config.ProjectConfig) error {
	return h.docker.Stop(cfg)
}

// Shell routes into the Nix dev shell. The editor and language tooling need
// go/node/python on PATH, and those live in the nix shell — not in the
// service container. args are forwarded as a one-shot command when set.
func (h *HybridProvider) Shell(cfg *config.ProjectConfig, args []string) error {
	return h.nix.Shell(cfg, args)
}

// Status reports both legs. Running tracks containers; nix is an ambient
// capability and never "down" on its own. Errors from either leg are
// surfaced together without aborting the report.
func (h *HybridProvider) Status(cfg *config.ProjectConfig) (EnvironmentStatus, error) {
	dockerStatus, dockerErr := h.docker.Status(cfg)
	nixStatus, nixErr := h.nix.Status(cfg)

	details := []string{
		"docker: " + describeLeg(dockerStatus, dockerErr),
		"nix: " + describeLeg(nixStatus, nixErr),
	}

	return EnvironmentStatus{
		Running: dockerStatus.Running,
		Details: strings.Join(details, " | "),
	}, errors.Join(dockerErr, nixErr)
}

func describeLeg(s EnvironmentStatus, err error) string {
	if err != nil {
		return "error: " + err.Error()
	}
	if strings.TrimSpace(s.Details) == "" {
		if s.Running {
			return "running"
		}
		return "not running"
	}
	return strings.TrimSpace(s.Details)
}
