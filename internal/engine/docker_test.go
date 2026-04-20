package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const composeFixture = `services:
  api:
    image: alpine:3.19
    command: ["sleep", "60"]
  worker:
    image: alpine:3.19
    command: ["sleep", "60"]
`

func TestIsDockerInstalled(t *testing.T) {
	result := IsDockerInstalled()
	assert.IsType(t, true, result, "IsDockerInstalled should return a boolean value")
}

func TestFirstService_PreservesDeclarationOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docker-compose.yml")
	require.NoError(t, os.WriteFile(path, []byte(composeFixture), 0o644))

	svc, err := FirstService(path)
	require.NoError(t, err)
	assert.Equal(t, "api", svc, "first service in YAML declaration order should win")
}

func TestFirstService_ErrorsOnEmptyCompose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docker-compose.yml")
	require.NoError(t, os.WriteFile(path, []byte("services: {}\n"), 0o644))

	_, err := FirstService(path)
	require.Error(t, err, "a compose file with no services must error")
}

func TestGenerateNetworkOverride_LabelsEveryService(t *testing.T) {
	dir := t.TempDir()
	composePath := filepath.Join(dir, "docker-compose.yml")
	require.NoError(t, os.WriteFile(composePath, []byte(composeFixture), 0o644))

	outDir := filepath.Join(dir, ".derrick")
	overridePath, err := GenerateNetworkOverride(composePath, outDir)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(outDir, "docker-compose.override.yml"), overridePath)

	data, err := os.ReadFile(overridePath)
	require.NoError(t, err)

	var got OverrideMap
	require.NoError(t, yaml.Unmarshal(data, &got))

	// Both services defined in the fixture must receive the derrick-managed
	// label (so `derrick clean` scopes prune) and the host-gateway alias
	// (so services can reach the host).
	require.Contains(t, got.Services, "api")
	require.Contains(t, got.Services, "worker")
	for _, svc := range got.Services {
		assert.Equal(t, "true", svc.Labels["com.derrick.managed"])
		assert.Contains(t, svc.ExtraHosts, "host.docker.internal:host-gateway")
	}
}

func TestGenerateNetworkOverride_ReturnsErrorWhenComposeMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := GenerateNetworkOverride(filepath.Join(dir, "nope.yml"), filepath.Join(dir, ".derrick"))
	require.Error(t, err, "missing compose file should surface a read error")
}
