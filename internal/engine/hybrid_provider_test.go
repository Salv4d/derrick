package engine

import (
	"errors"
	"testing"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/stretchr/testify/assert"
)

// stubLeg is a providerLeg double that records calls and returns the canned
// results set on it. Using a stub here (instead of a real docker/nix
// provider) lets us assert the hybrid composition logic without spawning
// a daemon or touching the filesystem.
type stubLeg struct {
	name string

	availableErr error
	startErr     error
	stopErr      error
	shellErr     error
	status       EnvironmentStatus
	statusErr    error

	startCalls int
	stopCalls  int
	shellCalls int
	shellArgs  []string
}

func (s *stubLeg) Name() string         { return s.name }
func (s *stubLeg) IsAvailable() error   { return s.availableErr }
func (s *stubLeg) Start(_ *config.ProjectConfig, _ Flags) error {
	s.startCalls++
	return s.startErr
}
func (s *stubLeg) Stop(_ *config.ProjectConfig) error {
	s.stopCalls++
	return s.stopErr
}
func (s *stubLeg) Shell(_ *config.ProjectConfig, args []string) error {
	s.shellCalls++
	s.shellArgs = args
	return s.shellErr
}
func (s *stubLeg) Status(_ *config.ProjectConfig) (EnvironmentStatus, error) {
	return s.status, s.statusErr
}

func newHybridWithStubs(docker, nix *stubLeg) *HybridProvider {
	return &HybridProvider{docker: docker, nix: nix}
}

func TestHybrid_IsAvailable_BothOK(t *testing.T) {
	h := newHybridWithStubs(&stubLeg{name: "docker"}, &stubLeg{name: "nix"})
	assert.NoError(t, h.IsAvailable())
}

func TestHybrid_IsAvailable_JoinsBothFailures(t *testing.T) {
	h := newHybridWithStubs(
		&stubLeg{name: "docker", availableErr: errors.New("docker missing")},
		&stubLeg{name: "nix", availableErr: errors.New("nix missing")},
	)
	err := h.IsAvailable()
	assert.ErrorContains(t, err, "docker missing")
	assert.ErrorContains(t, err, "nix missing")
}

func TestHybrid_Start_StopsOnDockerFailure(t *testing.T) {
	docker := &stubLeg{name: "docker", startErr: errors.New("compose up failed")}
	nix := &stubLeg{name: "nix"}

	err := newHybridWithStubs(docker, nix).Start(&config.ProjectConfig{}, Flags{})
	assert.ErrorContains(t, err, "compose up failed")
	assert.Equal(t, 1, docker.startCalls, "docker.Start should still have been invoked once")
	assert.Equal(t, 0, nix.startCalls, "nix.Start must not run when docker fails")
}

func TestHybrid_Start_RunsNixAfterDocker(t *testing.T) {
	docker := &stubLeg{name: "docker"}
	nix := &stubLeg{name: "nix"}

	err := newHybridWithStubs(docker, nix).Start(&config.ProjectConfig{}, Flags{})
	assert.NoError(t, err)
	assert.Equal(t, 1, docker.startCalls)
	assert.Equal(t, 1, nix.startCalls)
}

func TestHybrid_Stop_OnlyStopsDocker(t *testing.T) {
	docker := &stubLeg{name: "docker"}
	nix := &stubLeg{name: "nix"}

	assert.NoError(t, newHybridWithStubs(docker, nix).Stop(&config.ProjectConfig{}))
	assert.Equal(t, 1, docker.stopCalls)
	assert.Equal(t, 0, nix.stopCalls, "nix.Stop must be a no-op in hybrid")
}

func TestHybrid_Shell_RoutesToNix(t *testing.T) {
	docker := &stubLeg{name: "docker"}
	nix := &stubLeg{name: "nix"}

	assert.NoError(t, newHybridWithStubs(docker, nix).Shell(&config.ProjectConfig{}, []string{"go", "version"}))
	assert.Equal(t, 0, docker.shellCalls, "shell should not exec into the container")
	assert.Equal(t, 1, nix.shellCalls)
	assert.Equal(t, []string{"go", "version"}, nix.shellArgs)
}

func TestHybrid_Status_AggregatesBothLegs(t *testing.T) {
	docker := &stubLeg{name: "docker", status: EnvironmentStatus{Running: true, Details: "web, db"}}
	nix := &stubLeg{name: "nix", status: EnvironmentStatus{Running: true, Details: "flake available"}}

	got, err := newHybridWithStubs(docker, nix).Status(&config.ProjectConfig{})
	assert.NoError(t, err)
	assert.True(t, got.Running, "running tracks the docker leg")
	assert.Contains(t, got.Details, "docker: web, db")
	assert.Contains(t, got.Details, "nix: flake available")
}

func TestHybrid_Status_ReportsBothErrorsWithoutAborting(t *testing.T) {
	docker := &stubLeg{name: "docker", statusErr: errors.New("daemon unreachable")}
	nix := &stubLeg{name: "nix", statusErr: errors.New("eval failed")}

	got, err := newHybridWithStubs(docker, nix).Status(&config.ProjectConfig{})
	// Both leg errors should be joined — this is the contract change from the
	// previous implementation where the first error short-circuited.
	assert.ErrorContains(t, err, "daemon unreachable")
	assert.ErrorContains(t, err, "eval failed")
	assert.Contains(t, got.Details, "error: daemon unreachable")
	assert.Contains(t, got.Details, "error: eval failed")
}
