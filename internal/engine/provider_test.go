package engine

import (
	"testing"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestResolveProvider(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.ProjectConfig
		wantName     string
		wantConcrete any
	}{
		{
			name:         "explicit docker",
			cfg:          &config.ProjectConfig{Provider: "docker"},
			wantName:     "docker",
			wantConcrete: &DockerProvider{},
		},
		{
			name:         "explicit nix",
			cfg:          &config.ProjectConfig{Provider: "nix"},
			wantName:     "nix",
			wantConcrete: &NixProvider{},
		},
		{
			name:         "explicit hybrid",
			cfg:          &config.ProjectConfig{Provider: "hybrid"},
			wantName:     "hybrid",
			wantConcrete: &HybridProvider{},
		},
		{
			name:         "auto + compose file → docker",
			cfg:          &config.ProjectConfig{Provider: "auto", Docker: config.DockerConfig{Compose: "docker-compose.yml"}},
			wantName:     "docker",
			wantConcrete: &DockerProvider{},
		},
		{
			name:         "auto + nix packages → nix",
			cfg:          &config.ProjectConfig{Provider: "auto", Nix: config.NixConfig{Packages: []config.NixPackage{{Name: "go"}}}},
			wantName:     "nix",
			wantConcrete: &NixProvider{},
		},
		{
			name:         "empty provider defaults to nix",
			cfg:          &config.ProjectConfig{},
			wantName:     "nix",
			wantConcrete: &NixProvider{},
		},
		{
			name:         "unknown provider falls back to nix",
			cfg:          &config.ProjectConfig{Provider: "wasm"},
			wantName:     "nix",
			wantConcrete: &NixProvider{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := ResolveProvider(tc.cfg)
			assert.Equal(t, tc.wantName, p.Name())
			assert.IsType(t, tc.wantConcrete, p)
		})
	}
}
