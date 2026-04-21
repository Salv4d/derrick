package cli_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Salv4d/derrick/internal/config"
)

// TestRecipes_Parse verifies every full derrick.yaml block embedded in the
// recipe docs parses cleanly against the current schema. "Full recipe" is
// identified by the first non-blank line starting with `name:` — hook and
// profile fragments that don't stand alone are skipped. This guards against
// schema drift in published recipes.
func TestRecipes_Parse(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	recipeGlob := filepath.Join(repoRoot, "website", "docs", "use_cases", "*.md")

	files, err := filepath.Glob(recipeGlob)
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("no recipe markdown files found at %s", recipeGlob)
	}

	fenceRe := regexp.MustCompile("(?s)```yaml\\s*\\n(.*?)\\n```")

	totalBlocks := 0
	for _, path := range files {
		rel, _ := filepath.Rel(repoRoot, path)
		if filepath.Base(path) == "index.md" {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}

		blocks := fenceRe.FindAllStringSubmatch(string(data), -1)
		fullRecipes := 0
		for i, m := range blocks {
			yamlBody := m[1]
			if !isFullRecipe(yamlBody) {
				continue
			}
			fullRecipes++
			totalBlocks++

			t.Run(rel+"/block"+strconv.Itoa(i), func(t *testing.T) {
				if _, err := config.ParseConfigBytes([]byte(yamlBody), ""); err != nil {
					t.Errorf("recipe failed to parse:\n%s\n\nerror: %v", yamlBody, err)
				}
			})
		}
		if fullRecipes == 0 {
			t.Errorf("%s contains no full derrick.yaml block (expected a yaml fence starting with `name:`)", rel)
		}
	}

	if totalBlocks == 0 {
		t.Fatal("no recipe yaml blocks were discovered — regexp or glob is broken")
	}
}

// isFullRecipe returns true when the first non-blank line starts with `name:`,
// which distinguishes a complete derrick.yaml from a hook/profile fragment.
func isFullRecipe(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		return strings.HasPrefix(trimmed, "name:")
	}
	return false
}

