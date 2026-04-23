package engine

import (
	"testing"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestShouldRun(t *testing.T) {
	tests := []struct {
		name string
		when config.Condition
		opts HookOpts
		want bool
	}{
		{"empty when is always", nil, HookOpts{}, true},
		{"explicit always", config.Condition{"always"}, HookOpts{SetupCompleted: true}, true},
		{"first-setup fires when setup incomplete", config.Condition{"first-setup"}, HookOpts{SetupCompleted: false}, true},
		{"first-setup skips after initial start", config.Condition{"first-setup"}, HookOpts{SetupCompleted: true}, false},
		{"flag:seed fires when flag is active", config.Condition{"flag:seed"}, HookOpts{ActiveFlags: map[string]bool{"seed": true}}, true},
		{"flag:seed skips when flag is inactive", config.Condition{"flag:seed"}, HookOpts{ActiveFlags: map[string]bool{"seed": false}}, false},
		{"multi condition (OR logic): first-setup matches", config.Condition{"first-setup", "flag:reinstall"}, HookOpts{SetupCompleted: false}, true},
		{"multi condition (OR logic): flag matches", config.Condition{"first-setup", "flag:reinstall"}, HookOpts{SetupCompleted: true, ActiveFlags: map[string]bool{"reinstall": true}}, true},
		{"multi condition (OR logic): neither match", config.Condition{"first-setup", "flag:reinstall"}, HookOpts{SetupCompleted: true, ActiveFlags: map[string]bool{"reinstall": false}}, false},
		{"unknown when falls back to always", config.Condition{"some-other-mode"}, HookOpts{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, shouldRun(tc.when, tc.opts))
		})
	}
}
