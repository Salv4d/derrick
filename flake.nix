{
  description = "Derrick — local development environment orchestrator";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = "0.1.0-dev";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "derrick";
          inherit version;
          src = ./.;
          vendorHash = "sha256-gJVrkgrvFEB2OkYdMut1T6UBNj8qOv5rjmEa7mYpDJ8=";

          subPackages = [ "cmd/derrick" ];

          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${version}"
          ];

          meta = with pkgs.lib; {
            description = "Local development environment orchestrator (Nix + Docker)";
            homepage = "https://github.com/Salv4d/derrick";
            license = licenses.mit;
            mainProgram = "derrick";
          };
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go gofumpt golangci-lint ];
        };
      });
}
