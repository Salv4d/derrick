package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldRun(t *testing.T) {
	tests := []struct {
		name string
		when string
		opts HookOpts
		want bool
	}{
		{"empty when is always", "", HookOpts{}, true},
		{"explicit always", "always", HookOpts{SetupCompleted: true}, true},
		{"first-setup fires when setup incomplete", "first-setup", HookOpts{SetupCompleted: false}, true},
		{"first-setup skips after initial start", "first-setup", HookOpts{SetupCompleted: true}, false},
		{"flag:seed fires when flag is active", "flag:seed", HookOpts{ActiveFlags: map[string]bool{"seed": true}}, true},
		{"flag:seed skips when flag is inactive", "flag:seed", HookOpts{ActiveFlags: map[string]bool{"seed": false}}, false},
		{"flag:seed skips when flag is absent", "flag:seed", HookOpts{ActiveFlags: map[string]bool{}}, false},
		{"unknown when falls back to always", "some-other-mode", HookOpts{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, shouldRun(tc.when, tc.opts))
		})
	}
}
