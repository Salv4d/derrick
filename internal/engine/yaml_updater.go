package engine

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Salv4d/derrick/internal/ui"
)

func UpdateYAMLPackage(oldPkg, newPkg string) error {
	path := "derrick.yaml"
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Could not read derrick.yaml for auto-update: %w", err)
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
		ui.Warningf("Could not auto-update '%s' in derrick.yaml. Please update it manually.", oldPkg)
		return nil
	}

	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(path, []byte(newContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write updated derrick.yaml: %w", err)
	}

	ui.Successf("Auto-updated derrick.yaml: %s → %s", oldPkg, newPkg)
	return nil
}

func UpdateYAMLRegistry(newRegistry string) error {
	path := "derrick.yaml"
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read derrick.yaml for auto-update: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	// Regex to match an existing registry line and preserve its whitespace/comments
	// Group 1: Leading space and key -> "  nix_registry: "
	// Group 2: Trailing comments -> " # Legacy requirement"
	re := regexp.MustCompile(`^(\s*nix_registry:\s*)["']?[^"'\s#]+["']?(.*)$`)

	inDependencies := false
	registryReplaced := false
	dependenciesIdx := -1

	// Pass 1: Parse the file to find the block and update if the key exists
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 1. Detect the start of the dependencies block
		if strings.HasPrefix(trimmed, "dependencies:") {
			inDependencies = true
			dependenciesIdx = i // Remember this line number in case we need to inject later!
			newLines = append(newLines, line)
			continue
		}

		// 2. Detect exiting the block (Any new top-level key with no indentation)
		if inDependencies && trimmed != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(trimmed, "#") {
			inDependencies = false
		}

		// 3. If we are inside the block and find the key, execute the regex replacement
		if inDependencies && strings.HasPrefix(trimmed, "nix_registry:") && !registryReplaced {
			newLine := re.ReplaceAllString(line, fmt.Sprintf(`${1}"%s"${2}`, newRegistry))
			newLines = append(newLines, newLine)
			registryReplaced = true
			continue
		}

		newLines = append(newLines, line)
	}

	// Pass 2: The Injection (If the key didn't exist at all)
	if !registryReplaced {
		if dependenciesIdx == -1 {
			return fmt.Errorf("could not find 'dependencies:' block in derrick.yaml")
		}

		// Create a brand new line with standard 2-space YAML indentation
		injectedLine := fmt.Sprintf(`  nix_registry: "%s"`, newRegistry)

		// A classic Go slice manipulation to insert an item into the middle of an array
		newLines = append(newLines[:dependenciesIdx+1], append([]string{injectedLine}, newLines[dependenciesIdx+1:]...)...)
	}

	// Write the updated lines back to disk
	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated derrick.yaml: %w", err)
	}

	ui.Successf("Pinned Derrick to legacy Nix registry: %s", newRegistry)
	return nil
}