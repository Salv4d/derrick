package discovery

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectMetadata holds information about a detected project.
type ProjectMetadata struct {
	Name     string
	Version  string
	// Language is the detected programming language (e.g. "node", "go", "python").
	Language string
}

// LanguageNixSuggestions maps a detected language to a recommended set of Nix packages.
var LanguageNixSuggestions = map[string][]string{
	// nodejs includes npm — nodePackages namespace was removed from nixpkgs
	"node": {"nodejs"},
	// gopls is the Go language server, useful for IDE support
	"go": {"go", "gopls"},
	// python3 in nixpkgs includes pip; python3Packages.pip is a separate wrapper
	"python": {"python3"},
	// rust-analyzer is the LSP; available directly in nixpkgs unstable
	"rust": {"rustc", "cargo", "rust-analyzer"},
	// composer is available as a top-level package in nixpkgs
	"php": {"php", "phpPackages.composer"},
	// jdk21 is current LTS; maven as standard build tool
	"java": {"jdk21", "maven"},
	// bundler is available as a top-level nixpkgs package
	"ruby": {"ruby", "bundler"},
	// dotnet-sdk covers C# / F# development
	"csharp": {"dotnet-sdk"},
	// gcc + cmake + gnumake cover most C/C++ build setups
	"cpp": {"gcc", "cmake", "gnumake"},
	// swift toolchain is available in nixpkgs unstable
	"swift": {"swift"},
}

// SuggestedPackages returns the recommended Nix packages for the detected language,
// or an empty slice when the language is unknown.
func SuggestedPackages(lang string) []string {
	if pkgs, ok := LanguageNixSuggestions[lang]; ok {
		return pkgs
	}
	return []string{}
}

// Detector defines the contract for project type detectors.
type Detector interface {
	Detect(dir string) (*ProjectMetadata, bool)
}

// DiscoverProject runs all detectors to identify the project type and metadata.
func DiscoverProject(dir string) *ProjectMetadata {
	detectors := []Detector{
		&NodeDetector{},
		&GoDetector{},
		&PythonDetector{},
		&RustDetector{},
		&PHPDetector{},
		&JavaMavenDetector{},
		&JavaGradleDetector{},
		&RubyDetector{},
		&CSharpDetector{},
		&CppCMakeDetector{},
		&SwiftDetector{},
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	for _, d := range detectors {
		if meta, ok := d.Detect(dir); ok {
			if meta.Name == "" {
				meta.Name = sanitizeName(filepath.Base(absDir))
			}
			if meta.Version == "" {
				meta.Version = "0.1.0"
			}
			return meta
		}
	}

	return &ProjectMetadata{
		Name:    sanitizeName(filepath.Base(absDir)),
		Version: "0.1.0",
	}
}

func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

func extractRegex(content string, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// NodeDetector identifies Node.js projects via package.json.
type NodeDetector struct{}

func (d *NodeDetector) Detect(dir string) (*ProjectMetadata, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, false
	}
	var pkg struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	_ = json.Unmarshal(b, &pkg)
	return &ProjectMetadata{Name: pkg.Name, Version: pkg.Version, Language: "node"}, true
}

// GoDetector identifies Go projects via go.mod.
type GoDetector struct{}

func (d *GoDetector) Detect(dir string) (*ProjectMetadata, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil, false
	}
	name := extractRegex(string(b), `(?m)^module\s+([^\s]+)`)
	if name != "" {
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}
	return &ProjectMetadata{Name: name, Version: "0.1.0", Language: "go"}, true
}

// PythonDetector identifies Python projects via pyproject.toml.
type PythonDetector struct{}

func (d *PythonDetector) Detect(dir string) (*ProjectMetadata, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
	if err == nil {
		name := extractRegex(string(b), `(?m)name\s*=\s*["']([^"']+)["']`)
		version := extractRegex(string(b), `(?m)version\s*=\s*["']([^"']+)["']`)
		return &ProjectMetadata{Name: name, Version: version, Language: "python"}, true
	}
	b, err = os.ReadFile(filepath.Join(dir, "setup.py"))
	if err == nil {
		name := extractRegex(string(b), `(?m)name\s*=\s*["']([^"']+)["']`)
		version := extractRegex(string(b), `(?m)version\s*=\s*["']([^"']+)["']`)
		return &ProjectMetadata{Name: name, Version: version, Language: "python"}, true
	}
	return nil, false
}

// RustDetector identifies Rust projects via Cargo.toml.
type RustDetector struct{}

func (d *RustDetector) Detect(dir string) (*ProjectMetadata, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
	if err != nil {
		return nil, false
	}
	name := extractRegex(string(b), `(?m)name\s*=\s*["']([^"']+)["']`)
	version := extractRegex(string(b), `(?m)version\s*=\s*["']([^"']+)["']`)
	return &ProjectMetadata{Name: name, Version: version, Language: "rust"}, true
}

// PHPDetector identifies PHP projects via composer.json.
type PHPDetector struct{}

func (d *PHPDetector) Detect(dir string) (*ProjectMetadata, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "composer.json"))
	if err != nil {
		return nil, false
	}
	var pkg struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	_ = json.Unmarshal(b, &pkg)
	if pkg.Name != "" {
		parts := strings.Split(pkg.Name, "/")
		pkg.Name = parts[len(parts)-1]
	}
	return &ProjectMetadata{Name: pkg.Name, Version: pkg.Version, Language: "php"}, true
}

// JavaMavenDetector identifies Java Maven projects via pom.xml.
type JavaMavenDetector struct{}

func (d *JavaMavenDetector) Detect(dir string) (*ProjectMetadata, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "pom.xml"))
	if err != nil {
		return nil, false
	}
	name := extractRegex(string(b), `(?m)<artifactId>([^<]+)</artifactId>`)
	version := extractRegex(string(b), `(?m)<version>([^<]+)</version>`)
	return &ProjectMetadata{Name: name, Version: version, Language: "java"}, true
}

// JavaGradleDetector identifies Java Gradle projects via build.gradle or build.gradle.kts.
type JavaGradleDetector struct{}

func (d *JavaGradleDetector) Detect(dir string) (*ProjectMetadata, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "build.gradle"))
	if err != nil {
		b, err = os.ReadFile(filepath.Join(dir, "build.gradle.kts"))
		if err != nil {
			return nil, false
		}
	}
	version := extractRegex(string(b), `(?m)version\s*=?\s*["']([^"']+)["']`)
	name := ""
	s, err := os.ReadFile(filepath.Join(dir, "settings.gradle"))
	if err == nil {
		name = extractRegex(string(s), `(?m)rootProject\.name\s*=?\s*["']([^"']+)["']`)
	} else {
		s, err = os.ReadFile(filepath.Join(dir, "settings.gradle.kts"))
		if err == nil {
			name = extractRegex(string(s), `(?m)rootProject\.name\s*=?\s*["']([^"']+)["']`)
		}
	}
	return &ProjectMetadata{Name: name, Version: version, Language: "java"}, true
}

// RubyDetector identifies Ruby projects via Gemfile and gemspec.
type RubyDetector struct{}

func (d *RubyDetector) Detect(dir string) (*ProjectMetadata, bool) {
	if _, err := os.Stat(filepath.Join(dir, "Gemfile")); err != nil {
		return nil, false
	}
	var name, version string
	_ = filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".gemspec") {
			b, err := os.ReadFile(path)
			if err == nil {
				name = extractRegex(string(b), `(?m)name\s*=\s*["']([^"']+)["']`)
				version = extractRegex(string(b), `(?m)version\s*=\s*["']([^"']+)["']`)
			}
			return fs.SkipAll
		}
		if info.IsDir() && dir != path {
			return fs.SkipDir
		}
		return nil
	})
	return &ProjectMetadata{Name: name, Version: version, Language: "ruby"}, true
}

// CSharpDetector identifies C# projects via .csproj files.
type CSharpDetector struct{}

func (d *CSharpDetector) Detect(dir string) (*ProjectMetadata, bool) {
	var name string
	found := false
	_ = filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".csproj") {
			found = true
			b, err := os.ReadFile(path)
			if err == nil {
				name = extractRegex(string(b), `(?m)<AssemblyName>([^<]+)</AssemblyName>`)
				if name == "" {
					name = strings.TrimSuffix(info.Name(), ".csproj")
				}
			}
			return fs.SkipAll
		}
		if info.IsDir() && dir != path {
			return fs.SkipDir
		}
		return nil
	})
	if found {
		return &ProjectMetadata{Name: name, Version: "0.1.0", Language: "csharp"}, true
	}
	return nil, false
}

// CppCMakeDetector identifies C/C++ projects with CMake via CMakeLists.txt.
type CppCMakeDetector struct{}

func (d *CppCMakeDetector) Detect(dir string) (*ProjectMetadata, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "CMakeLists.txt"))
	if err != nil {
		return nil, false
	}
	name := extractRegex(string(b), `(?i)project\s*\(\s*([^ \)]+)`)
	version := extractRegex(string(b), `(?i)VERSION\s+([^\s\)]+)`)
	return &ProjectMetadata{Name: name, Version: version, Language: "cpp"}, true
}

// SwiftDetector identifies Swift projects via Package.swift.
type SwiftDetector struct{}

func (d *SwiftDetector) Detect(dir string) (*ProjectMetadata, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "Package.swift"))
	if err != nil {
		return nil, false
	}
	name := extractRegex(string(b), `(?m)name:\s*["']([^"']+)["']`)
	return &ProjectMetadata{Name: name, Version: "0.1.0", Language: "swift"}, true
}
