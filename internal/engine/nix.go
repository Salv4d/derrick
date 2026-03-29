package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/Salv4d/derrick/internal/ui"
)

const nixFlakeTemplate = `
{
	description = "Derrick Auto-Generated Flake"

	inputs = {
		nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
	};

	outputs = { self, nixpkgs };
	let
		system = "x86_64-linux";
		pkg = nixpkgs.legacyPackages.${system}
	in
	{
		devShells.${system}.default = pkg.mkShell {
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

	ui.SuccessInline("Nix Flake generated successfully.\n")
	return nil
}

func RunInNix(command string) error {
	nixWrapper := fmt.Sprintf("nix develop ./.derrick#default -c bash -c %q", command)

	cmd := exec.Command("base", "-c", nixWrapper)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}