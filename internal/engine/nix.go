package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

const nixFlakeTemplate = `
{
	description = "Derrick Auto-Generated Flake";

	inputs = {
		nixpkgs.url = "{{ .Registry }}";
	};

	outputs = { self, nixpkgs }:
	let
		system = "x86_64-linux";
		pkgs = nixpkgs.legacyPackages.${system};
	in
	{
		devShells.${system}.default = pkgs.mkShell {
			packages = with pkgs; [
			{{ range .Packages }}
			{{ . }}
			{{ end }}
			];
		};
	};
}
`

type NixTemplateData struct {
	Registry string
	Packages []string
}

func BootEnvironment(requestPackages []string, registryURL string) error {
	ui.Section("Derrick Sandbox Initialization")

	err := EnsureNixEnvironment(requestPackages, registryURL)
	if err != nil {
		return err
	}

	_, err = ValidateAndResolve(requestPackages, registryURL)
	if err != nil {
		return err
	}

	ui.Success("Environment verified and locked. Ready for execution.")
	return nil
}

func EnsureNixEnvironment(packages []string, customRegistry string) error {
	if len(packages) == 0 {
		return nil
	}

	ui.Task("Ensuring Nix environment isolation")

	registry := customRegistry
	if registry == "" {
		registry = config.DefaultNixRegistry
	}

	dir := ".derrick"
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		ui.Error("FAILED")
		return fmt.Errorf("failed to create %s directory: %w", dir, err)
	}

	tmpl, err := template.New("flake").Parse(nixFlakeTemplate)
	if err != nil {
		ui.Error("FAILED")
		return fmt.Errorf("failed to parse Nix template: %w", err)
	}

	var flakeContent bytes.Buffer
	data := NixTemplateData{Registry: registry, Packages: packages}
	err = tmpl.Execute(&flakeContent, data)
	if err != nil {
		ui.Error("FAILED")
		return fmt.Errorf("failed to execute Nix template: %w", err)
	}

	flakePath := filepath.Join(dir, "flake.nix")
	err = os.WriteFile(flakePath, flakeContent.Bytes(), 0o644)
	if err != nil {
		ui.Error("FAILED")
		return fmt.Errorf("failed to write flake.nix: %w", err)
	}

	ui.Success("DONE")
	return nil
}

func WrapWithNix(command string) []string {
	absPath, _ := filepath.Abs(".derrick")
	return []string{
		"nix",
		"develop",
		fmt.Sprintf("path:%s#default", absPath),
		"-c",
		"bash",
		"-c",
		command,
	}
}

func IsNixInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}
