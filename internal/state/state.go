package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const stateFile = ".derrick/state.json"

// Status represents the operational state of an environment.
type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusUnknown Status = "unknown"
)

// EnvironmentState captures everything Derrick knows about a project's runtime state.
// It is written to .derrick/state.json after a successful `derrick start`.
type EnvironmentState struct {
	Project             string    `json:"project"`
	Provider            string    `json:"provider"`
	Status              Status    `json:"status"`
	FirstSetupCompleted bool      `json:"first_setup_completed"`
	StartedAt           time.Time `json:"started_at,omitempty"`
	StoppedAt           time.Time `json:"stopped_at,omitempty"`
	ContainerIDs        []string  `json:"container_ids,omitempty"`
	FlagsUsed           []string  `json:"flags_used,omitempty"`
}

// Load reads the state file from the project directory.
// Returns a zeroed state (not an error) when no state file exists yet.
func Load(projectDir string) (*EnvironmentState, error) {
	path := filepath.Join(projectDir, stateFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &EnvironmentState{Status: StatusUnknown}, nil
	}
	if err != nil {
		return nil, err
	}

	var s EnvironmentState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Save writes the state to .derrick/state.json inside the project directory.
func Save(projectDir string, s *EnvironmentState) error {
	dir := filepath.Join(projectDir, ".derrick")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(projectDir, stateFile), data, 0o644)
}

// IsFirstSetup returns true when no successful start has been recorded.
func IsFirstSetup(projectDir string) bool {
	s, err := Load(projectDir)
	if err != nil {
		return true
	}
	return !s.FirstSetupCompleted
}
