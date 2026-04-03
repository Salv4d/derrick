package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/charmbracelet/huh"
)

type NixSearchRecord struct {
	PName       string `json:"pname"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

var undefinedRegex = regexp.MustCompile(`undefined variable '(.*?)'`)

func ValidateAndResolve(configPath string, packages []config.NixPackage, registryURL string, outDir string) ([]config.NixPackage, error) {
	if outDir == "" {
		outDir = ".derrick"
	}
	absPath, _ := filepath.Abs(outDir)

	cmd := exec.Command("nix", "develop", "--impure", fmt.Sprintf("path:%s#default", absPath), "-c", "true")
	cmd.Env = NixEnv()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		matches := undefinedRegex.FindStringSubmatch(errStr)

		if len(matches) > 1 {
			missingPkg := matches[1]

			if legacyRegistry, found := LegacyRegistryMap[missingPkg]; found {
				ui.Warningf("'%s' was removed from the unstable registry.", missingPkg)

				var useTimeMachine bool
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title(fmt.Sprintf("Derrick found it in a legacy snapshot. Pin registry to %s?", legacyRegistry)).
							Value(&useTimeMachine),
					),
				)
				if err := form.Run(); err == nil && useTimeMachine {
					_ = UpdateYAMLRegistry(configPath, legacyRegistry)
					ui.Successf("Pinned registry to legacy snapshot for '%s'.", missingPkg)

					EnsureNixEnvironment(configPath, packages, legacyRegistry, outDir)
					return ValidateAndResolve(configPath, packages, legacyRegistry, outDir)
				}
			}

			ui.Errorf("Nix package '%s' not found.", missingPkg)

			alternatives, searchErr := searchAlternatives(missingPkg)
			if searchErr != nil {
				return packages, fmt.Errorf("failed to search alternatives: %w", searchErr)
			}

			if len(alternatives) == 0 {
				return packages, fmt.Errorf("no similar packages found for '%s'. Please check your %s.", missingPkg, configPath)
			}

			var selectedPkg string
			options := make([]huh.Option[string], len(alternatives)+1)
			for i, alt := range alternatives {
				label := fmt.Sprintf("%s (v%s) - %s", alt.PName, alt.Version, truncate(alt.Description, 50))
				options[i] = huh.NewOption(label, alt.PName)
			}

			options[len(alternatives)] = huh.NewOption("✖ None of these (Abort)", "ABORT_RESOLUTION")

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title(fmt.Sprintf("Select alternative for '%s':", missingPkg)).
						Options(options...).
						Value(&selectedPkg),
				),
			)

			if err := form.Run(); err != nil {
				return packages, fmt.Errorf("resolution cancelled: %w", err)
			}

			if selectedPkg == "ABORT_RESOLUTION" {
				return packages, fmt.Errorf("user aborted package resolution. Please manually fix '%s' in %s", missingPkg, configPath)
			}

			for i, p := range packages {
				if p.Name == missingPkg {
					packages[i].Name = selectedPkg
					break
				}
			}

			_ = UpdateYAMLPackage(configPath, missingPkg, selectedPkg)
			ui.Successf("Resolved '%s' -> '%s'.", missingPkg, selectedPkg)

			EnsureNixEnvironment(configPath, packages, registryURL, outDir)
			return ValidateAndResolve(configPath, packages, registryURL, outDir)
		}

		return packages, fmt.Errorf("nix evaluation failed:\n%s\n\nRun 'derrick shell --debug' to investigate", errStr)
	}

	return packages, nil
}

func searchAlternatives(pkgName string) ([]NixSearchRecord, error) {
	ui.Taskf("Searching alternatives for '%s'", pkgName)

	results, err := executeNixSearch(pkgName)
	if err != nil {
		ui.Error("FAILED")
		return nil, err
	}

	if len(results) == 0 {
		re := regexp.MustCompile(`^([a-zA-Z]+[0-9]?)`)
		if match := re.FindStringSubmatch(pkgName); len(match) > 1 {
			baseName := match[1]
			if baseName != pkgName {
				ui.Taskf("No exact match. Widening search to '%s'", baseName)
				results, err = executeNixSearch(baseName)
				if err != nil {
					ui.Error("FAILED")
					return nil, err
				}
			}
		}
	}

	ui.Successf("Found %d possibilities", len(results))

	sort.Slice(results, func(i, j int) bool {
		if len(results[i].PName) == len(results[j].PName) {
			return results[i].PName < results[j].PName
		}
		return len(results[i].PName) < len(results[j].PName)
	})

	if len(results) > 7 {
		results = results[:7]
	}

	return results, nil
}

func executeNixSearch(query string) ([]NixSearchRecord, error) {
	cmd := exec.Command("nix", "search", "nixpkgs", query, "--json")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var searchResults map[string]NixSearchRecord
	if err := json.Unmarshal(stdout.Bytes(), &searchResults); err != nil {
		return nil, err
	}

	var results []NixSearchRecord
	for _, record := range searchResults {
		results = append(results, record)
	}

	return results, nil
}

func truncate(text string, maxLen int) string {
	if len(text) > maxLen {
		return text[:maxLen] + "..."
	}
	return text
}
