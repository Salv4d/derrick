package engine

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Salv4d/derrick/internal/ui"
)

// UpdateYAMLPackage replaces a Nix package name under nix.packages in derrick.yaml.
func UpdateYAMLPackage(configPath, oldPkg, newPkg string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("could not read %s for auto-update: %w", configPath, err)
	}

	re := regexp.MustCompile(fmt.Sprintf(`^(\s*-\s*)["']?%s["']?\s*$`, regexp.QuoteMeta(oldPkg)))

	lines := strings.Split(string(content), "\n")
	var out []string
	inNix, inPkgs, replaced := false, false, false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		topLevel := len(line) > 0 && line[0] != ' ' && line[0] != '\t'

		switch {
		case trimmed == "nix:":
			inNix, inPkgs = true, false
		case inNix && topLevel && trimmed != "":
			inNix, inPkgs = false, false
		case inNix && trimmed == "packages:":
			inPkgs = true
		case inNix && inPkgs && trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "#"):
			inPkgs = false
		}

		if inPkgs && !replaced {
			if m := re.ReplaceAllString(line, fmt.Sprintf(`${1}"%s"`, newPkg)); m != line {
				line, replaced = m, true
			}
		}

		out = append(out, line)
	}

	if !replaced {
		ui.Warningf("Could not auto-update '%s' in %s. Please update it manually.", oldPkg, configPath)
		return nil
	}

	if err := os.WriteFile(configPath, []byte(strings.Join(out, "\n")), 0o644); err != nil {
		return fmt.Errorf("failed to write updated %s: %w", configPath, err)
	}
	ui.Successf("Auto-updated %s: %s → %s", configPath, oldPkg, newPkg)
	return nil
}

// UpdateYAMLRegistry sets or inserts nix.registry in derrick.yaml.
func UpdateYAMLRegistry(configPath, newRegistry string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("could not read %s for auto-update: %w", configPath, err)
	}

	registryRe := regexp.MustCompile(`^\s+registry:.*$`)

	lines := strings.Split(string(content), "\n")
	var out []string
	inNix, replaced := false, false
	nixIdx := -1

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		topLevel := len(line) > 0 && line[0] != ' ' && line[0] != '\t'

		if trimmed == "nix:" {
			inNix = true
			nixIdx = len(out)
		} else if inNix && topLevel && trimmed != "" {
			inNix = false
		}

		if inNix && registryRe.MatchString(line) && !replaced {
			out = append(out, fmt.Sprintf(`  registry: "%s"`, newRegistry))
			replaced = true
			continue
		}

		out = append(out, line)
	}

	if !replaced {
		if nixIdx == -1 {
			out = append(out, "nix:")
			out = append(out, fmt.Sprintf(`  registry: "%s"`, newRegistry))
		} else {
			insert := []string{fmt.Sprintf(`  registry: "%s"`, newRegistry)}
			out = append(out[:nixIdx+1], append(insert, out[nixIdx+1:]...)...)
		}
	}

	if err := os.WriteFile(configPath, []byte(strings.Join(out, "\n")), 0o644); err != nil {
		return fmt.Errorf("failed to write updated %s: %w", configPath, err)
	}
	ui.Successf("Pinned nix.registry in %s: %s", configPath, newRegistry)
	return nil
}
