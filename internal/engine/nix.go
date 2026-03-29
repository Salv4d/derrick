package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/Salv4d/derrick/internal/ui"
)

const nixFlakeTemplate = `
{
	description = "Derrick Auto-Generated Flake";

	inputs = {
		nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
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
	Packages []string
}

func EnsureNixEnvironment(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	ui.Info("Ensuring Nix environment is strictly isolated...")

	dir := ".derrick"
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create %s directory: %w", dir, err)
	}

	tmpl, err := template.New("flake").Parse(nixFlakeTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse Nix template: %w", err)
	}

	var flakeContent bytes.Buffer
	data := NixTemplateData{Packages: packages}
	err = tmpl.Execute(&flakeContent, data)
	if err != nil {
		return fmt.Errorf("failed to execute Nix template: %w", err)
	}

	flakePath := filepath.Join(dir, "flake.nix")
	err = os.WriteFile(flakePath, flakeContent.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write flake.nix: %w", err)
	}

	ui.SuccessInline("Nix Flake generated successfully.\n")
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