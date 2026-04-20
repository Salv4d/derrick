//go:build integration

// Integration tests exercise a real docker daemon.
// Run with: go test -tags=integration -count=1 ./tests/...
package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dockerFixtureYAML = `name: "derrick-it"
version: "0.0.1"
provider: docker

docker:
  compose: "docker-compose.yml"
  shell: "app"

hooks:
  setup:
    - run: "echo setting-up"
      when: always
  after_start:
    - run: "echo started"
      when: always
  before_stop:
    - run: "echo stopping"
      when: always
`

const dockerFixtureCompose = `services:
  app:
    image: alpine:3.19
    command: ["sleep", "60"]
`

// buildDerrick compiles the CLI once per test binary and returns its path.
func buildDerrick(t *testing.T) string {
	t.Helper()
	repoRoot, err := filepath.Abs("..")
	require.NoError(t, err)

	bin := filepath.Join(t.TempDir(), "derrick")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/derrick")
	cmd.Dir = repoRoot
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	require.NoErrorf(t, cmd.Run(), "go build failed: %s", buf.String())
	return bin
}

func TestIntegration_DockerStartStop(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}

	bin := buildDerrick(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "derrick.yaml"), []byte(dockerFixtureYAML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(dockerFixtureCompose), 0o644))

	derrick := func(args ...string) (string, error) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = dir
		var buf bytes.Buffer
		cmd.Stdout, cmd.Stderr = &buf, &buf
		runErr := cmd.Run()
		return buf.String(), runErr
	}

	t.Cleanup(func() {
		_, _ = derrick("stop")
		_ = exec.Command("docker", "compose", "-f", filepath.Join(dir, "docker-compose.yml"), "down", "-v", "--remove-orphans").Run()
	})

	out, err := derrick("start")
	require.NoErrorf(t, err, "derrick start failed: %s", out)

	deadline := time.Now().Add(20 * time.Second)
	var ps string
	for time.Now().Before(deadline) {
		raw, _ := exec.Command("docker", "ps", "--format", "{{.Names}}").Output()
		ps = string(raw)
		if strings.Contains(ps, "derrick-it") {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	assert.Containsf(t, ps, "derrick-it", "expected a derrick-it container running\nderrick start output:\n%s", out)

	stopOut, err := derrick("stop")
	require.NoErrorf(t, err, "derrick stop failed: %s", stopOut)

	raw, _ := exec.Command("docker", "ps", "--format", "{{.Names}}").Output()
	assert.NotContains(t, string(raw), "derrick-it", "container should be torn down after stop")
}
