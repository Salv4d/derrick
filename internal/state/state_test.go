package state

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()

	// Load from a directory with no state file — should return zeroed state, not an error.
	s, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, StatusUnknown, s.Status)
	assert.False(t, s.FirstSetupCompleted)

	// Mutate and save.
	s.Project = "my-api"
	s.Provider = "docker"
	s.Status = StatusRunning
	s.FirstSetupCompleted = true
	s.StartedAt = time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	s.FlagsUsed = []string{"seed-db"}

	require.NoError(t, Save(dir, s))

	// Load again and verify the round-trip.
	loaded, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, "my-api", loaded.Project)
	assert.Equal(t, "docker", loaded.Provider)
	assert.Equal(t, StatusRunning, loaded.Status)
	assert.True(t, loaded.FirstSetupCompleted)
	assert.Equal(t, []string{"seed-db"}, loaded.FlagsUsed)
}

func TestIsFirstSetup(t *testing.T) {
	dir := t.TempDir()

	// No state file — should be first setup.
	assert.True(t, IsFirstSetup(dir))

	// Save a completed state.
	require.NoError(t, Save(dir, &EnvironmentState{
		FirstSetupCompleted: true,
		Status:              StatusRunning,
	}))

	assert.False(t, IsFirstSetup(dir))
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := dir + "/project/subdir"

	err := Save(nested, &EnvironmentState{Status: StatusStopped})
	require.NoError(t, err)

	_, err = os.Stat(nested + "/.derrick/state.json")
	assert.NoError(t, err, "state.json should have been created in nested directory")
}

func TestLoadCorruptedFile(t *testing.T) {
	dir := t.TempDir()
	stateDir := dir + "/.derrick"
	require.NoError(t, os.MkdirAll(stateDir, 0o755))
	require.NoError(t, os.WriteFile(stateDir+"/state.json", []byte("not json {{{"), 0o644))

	_, err := Load(dir)
	assert.Error(t, err, "corrupted JSON should return an error")
}
