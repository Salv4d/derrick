package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

const nixFlakeTemplate = `
{
	description = "Derrick Auto-Generated Flake";

	inputs = {
		nixpkgs.url = "{{ .Registry }}";{{ range $regName, $regUrl := .ExtraRegistries }}
		{{ $regName }}.url = "{{ $regUrl }}";{{ end }}
	};

	outputs = { self, nixpkgs{{ range $name, $url := .ExtraRegistries }}, {{ $name }}{{ end }}, ... }:
	let
		system = "{{ .System }}";
		pkgs = nixpkgs.legacyPackages.${system};{{ range $name, $url := .ExtraRegistries }}
		{{ $name }}_pkgs = {{ $name }}.legacyPackages.${system};{{ end }}
	in
	{
		devShells.${system}.default = pkgs.mkShell {
			packages = [
			{{ range .Packages }}
			{{ . }}
			{{ end }}
			];
			shellHook = ''
				export PS1='\[\e[34m\](derrick-sandbox)\[\e[0m\] \w > '
				export HISTFILE="$PWD/.derrick/shell_history"
				export HISTSIZE=10000
				export HISTFILESIZE=10000
				export HISTCONTROL=ignoredups:erasedups
			'';
		};
	};
}
`

// NixTemplateData holds data for rendering the Nix flake template.
type NixTemplateData struct {
	Registry        string
	System          string
	Packages        []string
	ExtraRegistries map[string]string
}

// nixSystem returns the nixpkgs system string for the current host,
// e.g. "x86_64-linux", "aarch64-darwin".
func nixSystem() string {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	}
	return fmt.Sprintf("%s-%s", arch, runtime.GOOS)
}

// BootEnvironment initializes and validates the Nix sandbox.
func BootEnvironment(configPath string, requestPackages []config.NixPackage, registryURL string, outDir string) error {
	ui.Section("Derrick Sandbox Initialization")

	updated, err := EnsureNixEnvironment(configPath, requestPackages, registryURL, outDir)
	if err != nil {
		return err
	}

	// Optimization: skip validation if flake.nix is unchanged and flake.lock exists.
	if !updated {
		lockPath := filepath.Join(outDir, "flake.lock")
		if outDir == "" {
			lockPath = ".derrick/flake.lock"
		}
		if _, err := os.Stat(lockPath); err == nil {
			ui.Success("Environment UP TO DATE. Skipping verification.")
			return nil
		}
	}

	_, err = ValidateAndResolve(configPath, requestPackages, registryURL, outDir)
	if err != nil {
		return err
	}

	ui.Success("Environment verified and locked. Ready for execution.")
	return nil
}

// EnsureNixEnvironment creates the flake.nix and ensures isolation.
// Returns true if flake.nix was updated.
func EnsureNixEnvironment(configPath string, packages []config.NixPackage, customRegistry string, outDir string) (bool, error) {
	if len(packages) == 0 {
		return false, nil
	}

	ui.Task("Ensuring Nix environment isolation")

	registry := customRegistry
	if registry == "" {
		registry = config.DefaultNixRegistry
	}

	dir := outDir
	if dir == "" {
		dir = ".derrick"
	}
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		ui.Error("FAILED")
		return false, fmt.Errorf("failed to create %s directory: %w", dir, err)
	}

	tmpl, err := template.New("flake").Parse(nixFlakeTemplate)
	if err != nil {
		ui.Error("FAILED")
		return false, fmt.Errorf("failed to parse Nix template: %w", err)
	}

	data := NixTemplateData{
		Registry:        registry,
		System:          nixSystem(),
		ExtraRegistries: make(map[string]string),
	}

	registryIndex := 1
	registryMap := make(map[string]string)

	for _, p := range packages {
		targetPkg := ""
		if p.Registry != "" && p.Registry != registry {
			regName, exists := registryMap[p.Registry]
			if !exists {
				regName = fmt.Sprintf("reg%d", registryIndex)
				registryIndex++
				registryMap[p.Registry] = regName
				data.ExtraRegistries[regName] = p.Registry
			}
			targetPkg = fmt.Sprintf("%s_pkgs.%s", regName, p.Name)
		} else {
			targetPkg = fmt.Sprintf("pkgs.%s", p.Name)
		}
		data.Packages = append(data.Packages, targetPkg)
	}

	var flakeContent bytes.Buffer
	err = tmpl.Execute(&flakeContent, data)
	if err != nil {
		ui.Error("FAILED")
		return false, fmt.Errorf("failed to execute Nix template: %w", err)
	}

	flakePath := filepath.Join(dir, "flake.nix")
	
	// Check if already exists and is identical
	if existing, err := os.ReadFile(flakePath); err == nil {
		if bytes.Equal(existing, flakeContent.Bytes()) {
			ui.Success("UP TO DATE")
			return false, nil
		}
	}

	err = os.WriteFile(flakePath, flakeContent.Bytes(), 0o644)
	if err != nil {
		ui.Error("FAILED")
		return false, fmt.Errorf("failed to write flake.nix: %w", err)
	}

	ui.Success("DONE")
	return true, nil
}

// NixEnv returns the current environment with NIXPKGS_ALLOW_UNFREE=1 set,
// so that all Nix invocations can resolve unfree packages (e.g. Cursor, VSCode).
// We also allow insecure packages since legacy snapshots inherently pull outdated versions.
func NixEnv() []string {
	return append(os.Environ(), "NIXPKGS_ALLOW_UNFREE=1", "NIXPKGS_ALLOW_INSECURE=1")
}

// WrapWithNix returns a command array to run inside the Nix environment.
func WrapWithNix(command string, outDir string) []string {
	if outDir == "" {
		outDir = ".derrick"
	}
	absPath, _ := filepath.Abs(outDir)
	return []string{
		"nix",
		"develop",
		"--impure",
		fmt.Sprintf("path:%s#default", absPath),
		"-c",
		"bash",
		"-c",
		command,
	}
}

// IsNixInstalled checks if the nix binary is available in PATH.
func IsNixInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// WriteEnvRC writes a direnv .envrc in projectDir that activates the Nix
// flake generated by derrick. It is a no-op if .envrc already exists.
// Returns true if the file was newly created.
func WriteEnvRC(projectDir string) (bool, error) {
	path := filepath.Join(projectDir, ".envrc")
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}
	const content = "# Generated by derrick — requires nix-direnv (https://github.com/nix-community/nix-direnv)\nuse flake ./.derrick#default\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return false, fmt.Errorf("failed to write .envrc: %w", err)
	}
	return true, nil
}
