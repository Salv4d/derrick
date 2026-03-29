package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/Salv4d/derrick/internal/ui"
	"github.com/charmbracelet/huh"
)

type NixSearchRecord struct {
	PName string `json:"pname"`
	Version string `json:"version"`
	Description string `json:"description"`
}

var undefinedRegex = regexp.MustCompile(`undefined variable '(.*?)'`)

func ValidateAndResolve(packages []string) ([]string, error) {
	absPath, _ := filepath.Abs(".derrick")

	cmd := exec.Command("nix", "develop", fmt.Sprintf("path:%s#default", absPath), "-c", "true")
	cmd.Env = append(os.Environ(), "NO_COLOR=1", "TERM=dumb")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		matches := undefinedRegex.FindStringSubmatch(errStr)

		if len(matches) > 1 {
			missingPkg := matches[1]
			ui.Errorf("Nix package '%s' not found in current registry.", missingPkg)

			alternatives, searchErr := searchAlternatives(missingPkg)
			if searchErr != nil {
				return packages, fmt.Errorf("failed to search alternatives: %w", searchErr)
			}

			if len(alternatives) == 0 {
				return packages, fmt.Errorf("no similar packages found for '%s'. Please check your derrick.yaml.", missingPkg)
			}

			var selectedPkg string
			options := make([]huh.Option[string], len(alternatives))
			for i, alt := range alternatives {
				label := fmt.Sprintf("%s (v%s) - %s", alt.PName, alt.Version, truncate(alt.Description, 50))
				options[i] = huh.NewOption(label, alt.PName)
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title(fmt.Sprintf("Did you mean one of these alternatives for '%s'?", missingPkg)).
						Options(options...).
						Value(&selectedPkg),
				),
			)

			if err := form.Run(); err != nil {
				return packages, fmt.Errorf("resolution cancelled: %w", err)
			}

			for i, p := range packages {
				if p == missingPkg {
					packages[i] = selectedPkg
					break
				}
			}

			ui.SuccessInlinef("Resolved '%s' to '%s'. Regenerating Flake...\n", missingPkg, selectedPkg)

			EnsureNixEnvironment(packages)
			return ValidateAndResolve(packages)
		}

		return packages, fmt.Errorf("nix evaluation failed:\n%s\n\nRun 'derrick shell --debug' to investigate", errStr)
	}

	return packages, nil
}

func searchAlternatives(pkgName string) ([]NixSearchRecord, error) {
	ui.Infof("Searching Nix registry for alternatives to '%s'...", pkgName)

	results, err := executeNixSearch(pkgName)
	if err != nil {
		return nil, err
	}
	
	if len(results) == 0 {
		re := regexp.MustCompile(`^([a-zA-Z]+[0-9]?)`)
		if match := re.FindStringSubmatch(pkgName); len(match) > 1 {
			baseName := match[1]
			if baseName != pkgName {
				ui.Infof("No exact binaries found. Widening search to '%s'...\n", baseName)
				results, err = executeNixSearch(baseName)
				if err != nil {
					return nil, err
				}
			}
		}
	}

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