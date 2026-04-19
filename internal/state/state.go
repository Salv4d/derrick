package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const (
	stateFile     = ".derrick/state.json"
	stateLockFile = ".derrick/state.lock"
)

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
	FlagsUsed           []string  `json:"flags_used,omitempty"`
}

// acquireLock opens (or creates) the lock file and acquires a syscall.Flock lock.
// The caller must call release() when the critical section ends.
func acquireLock(projectDir string, how int) (release func(), err error) {
	dir := filepath.Join(projectDir, ".derrick")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	lf, err := os.OpenFile(filepath.Join(projectDir, stateLockFile), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(lf.Fd()), how); err != nil {
		lf.Close()
		return nil, err
	}

	return func() {
		_ = syscall.Flock(int(lf.Fd()), syscall.LOCK_UN)
		lf.Close()
	}, nil
}

// Load reads the state file from the project directory under a shared read lock.
// Returns a zeroed state (not an error) when no state file exists yet.
func Load(projectDir string) (*EnvironmentState, error) {
	release, err := acquireLock(projectDir, syscall.LOCK_SH)
	if err != nil {
		return nil, err
	}
	defer release()

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

// Save writes the state to .derrick/state.json inside the project directory
// under an exclusive write lock, using an atomic rename to prevent partial writes.
func Save(projectDir string, s *EnvironmentState) error {
	release, err := acquireLock(projectDir, syscall.LOCK_EX)
	if err != nil {
		return err
	}
	defer release()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Join(projectDir, ".derrick")
	tmpPath := filepath.Join(dir, "state.json.tmp")
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, filepath.Join(projectDir, stateFile))
}

// IsFirstSetup returns true when no successful start has been recorded.
func IsFirstSetup(projectDir string) bool {
	s, err := Load(projectDir)
	if err != nil {
		return true
	}
	return !s.FirstSetupCompleted
}
