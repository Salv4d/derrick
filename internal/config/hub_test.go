package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadGlobalHub(t *testing.T) {
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Unsetenv("HOME")

	derrickDir := filepath.Join(tempHome, ".derrick")
	err := os.MkdirAll(derrickDir, 0o755)
	require.NoError(t, err)

	yamlContent := []byte(`
registries:
  - "https://raw.githubusercontent.com/Salv4d/derrick-registry/main/registry.json"
projects:
  backend-api: "git@github.com:Salv4d/backend-api.git"
  frontend-ui: "https://github.com/Salv4d/frontend-ui.git"
`)
	configPath := filepath.Join(derrickDir, "config.yaml")
	err = os.WriteFile(configPath, yamlContent, 0o644)
	require.NoError(t, err)

	hub, err := LoadGlobalHub()
	require.NoError(t, err)

	url, err := hub.ResolveAlias("backend-api")
	require.NoError(t, err)
	assert.Equal(t, "git@github.com:Salv4d/backend-api.git", url)

	url, err = hub.ResolveAlias("frontend-ui")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/Salv4d/frontend-ui.git", url)

	_, err = hub.ResolveAlias("missing-project")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoadGlobalHub_MissingFile(t *testing.T) {
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Unsetenv("HOME")

	hub, err := LoadGlobalHub()
	require.NoError(t, err, "Loading missing hub config should return empty struct, not error")
	assert.NotNil(t, hub.Projects)
	assert.Len(t, hub.Projects, 0)
}
