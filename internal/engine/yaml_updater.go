package engine

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Salv4d/derrick/internal/ui"
)

func UpdateYAMLPackage(configPath, oldPkg, newPkg string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Could not read %s for auto-update: %w", configPath, err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	pattern := fmt.Sprintf(`^(\s*-\s*)["']?%s["']?(.*)$`, regexp.QuoteMeta(oldPkg))
	re := regexp.MustCompile(pattern)

	inTargetBlock := false
	replaced := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "nix_packages:") {
			inTargetBlock = true
			newLines = append(newLines, line)
			continue
		}

		if inTargetBlock && trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "#") {
			inTargetBlock = false
		}

		if inTargetBlock && !replaced {
			newLine := re.ReplaceAllString(line, fmt.Sprintf(`${1}"%s"${2}`, newPkg))
			if newLine != line {
				line = newLine
				replaced = true
			}
		}

		newLines = append(newLines, line)
	}

	if !replaced {
		ui.Warningf("Could not auto-update '%s' in %s. Please update it manually.", oldPkg, configPath)
		return nil
	}

	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(configPath, []byte(newContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write updated %s: %w", configPath, err)
	}

	ui.Successf("Auto-updated %s: %s → %s", configPath, oldPkg, newPkg)
	return nil
}

func UpdateYAMLRegistry(configPath, newRegistry string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("could not read %s for auto-update: %w", configPath, err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	re := regexp.MustCompile(`^(\s*nix_registry:\s*)["']?[^"'\s#]+["']?(.*)$`)

	inDependencies := false
	registryReplaced := false
	dependenciesIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "dependencies:") {
			inDependencies = true
			dependenciesIdx = i
			newLines = append(newLines, line)
			continue
		}

		if inDependencies && trimmed != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(trimmed, "#") {
			inDependencies = false
		}

		if inDependencies && strings.HasPrefix(trimmed, "nix_registry:") && !registryReplaced {
			newLine := re.ReplaceAllString(line, fmt.Sprintf(`${1}"%s"${2}`, newRegistry))
			newLines = append(newLines, newLine)
			registryReplaced = true
			continue
		}

		newLines = append(newLines, line)
	}

	if !registryReplaced {
		if dependenciesIdx == -1 {
			return fmt.Errorf("could not find 'dependencies:' block in %s", configPath)
		}

		injectedLine := fmt.Sprintf(`  nix_registry: "%s"`, newRegistry)

		newLines = append(newLines[:dependenciesIdx+1], append([]string{injectedLine}, newLines[dependenciesIdx+1:]...)...)
	}

	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(configPath, []byte(newContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write updated %s: %w", configPath, err)
	}

	ui.Successf("Pinned Derrick to legacy Nix registry in %s: %s", configPath, newRegistry)
	return nil
}
