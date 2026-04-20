package engine

// LegacyEntry maps a package to the last nixpkgs channel that shipped it and,
// when the attribute was renamed between channels, the canonical name to use
// in that channel.
type LegacyEntry struct {
	Registry  string // pinned nixpkgs flake URL
	Attribute string // canonical attribute in that registry; empty = same as the lookup key
}

// LegacyPackages maps attribute names (as a user might write them, including
// common shorthands) to the last nixpkgs channel that shipped the package.
// When Attribute is set, derrick will also rename the entry in derrick.yaml
// so the pinned channel resolves it correctly.
var LegacyPackages = map[string]LegacyEntry{
	// ── Node.js ───────────────────────────────────────────────────────────
	// nixpkgs ≥24.05 uses nodejs_N; older channels used nodejs-N_x.
	// Both forms are indexed so either spelling resolves correctly.
	"nodejs_10":   {Registry: "github:NixOS/nixpkgs/nixos-19.09", Attribute: "nodejs-10_x"},
	"nodejs-10_x": {Registry: "github:NixOS/nixpkgs/nixos-19.09"},
	"nodejs_12":   {Registry: "github:NixOS/nixpkgs/nixos-21.05", Attribute: "nodejs-12_x"},
	"nodejs-12_x": {Registry: "github:NixOS/nixpkgs/nixos-21.05"},
	"nodejs_14":   {Registry: "github:NixOS/nixpkgs/nixos-22.05", Attribute: "nodejs-14_x"},
	"nodejs-14_x": {Registry: "github:NixOS/nixpkgs/nixos-22.05"},
	"nodejs_16":   {Registry: "github:NixOS/nixpkgs/nixos-23.05", Attribute: "nodejs-16_x"},
	"nodejs-16_x": {Registry: "github:NixOS/nixpkgs/nixos-23.05"},
	"nodejs_18":   {Registry: "github:NixOS/nixpkgs/nixos-24.05", Attribute: "nodejs-18_x"},
	"nodejs-18_x": {Registry: "github:NixOS/nixpkgs/nixos-24.05"},

	// ── Python ────────────────────────────────────────────────────────────
	"python36":  {Registry: "github:NixOS/nixpkgs/nixos-19.09"},
	"python37":  {Registry: "github:NixOS/nixpkgs/nixos-20.03"},
	"python38":  {Registry: "github:NixOS/nixpkgs/nixos-21.05"},
	"python39":  {Registry: "github:NixOS/nixpkgs/nixos-21.11"},
	"python310": {Registry: "github:NixOS/nixpkgs/nixos-22.11"},
	"python311": {Registry: "github:NixOS/nixpkgs/nixos-23.11"},

	// ── Ruby ──────────────────────────────────────────────────────────────
	"ruby_2_6": {Registry: "github:NixOS/nixpkgs/nixos-20.09"},
	"ruby_2_7": {Registry: "github:NixOS/nixpkgs/nixos-22.05"},
	"ruby_3_0": {Registry: "github:NixOS/nixpkgs/nixos-23.11"},

	// ── PHP ───────────────────────────────────────────────────────────────
	"php73": {Registry: "github:NixOS/nixpkgs/nixos-20.09"},
	"php74": {Registry: "github:NixOS/nixpkgs/nixos-22.05"},
	"php80": {Registry: "github:NixOS/nixpkgs/nixos-22.11"},
	"php81": {Registry: "github:NixOS/nixpkgs/nixos-23.11"},

	// ── Go ────────────────────────────────────────────────────────────────
	"go_1_17": {Registry: "github:NixOS/nixpkgs/nixos-21.11"},
	"go_1_18": {Registry: "github:NixOS/nixpkgs/nixos-22.11"},
	"go_1_19": {Registry: "github:NixOS/nixpkgs/nixos-23.05"},
	"go_1_20": {Registry: "github:NixOS/nixpkgs/nixos-23.11"},
	"go_1_21": {Registry: "github:NixOS/nixpkgs/nixos-24.05"},

	// ── Erlang / Elixir ───────────────────────────────────────────────────
	"erlangR24": {Registry: "github:NixOS/nixpkgs/nixos-22.05"},
	"erlangR25": {Registry: "github:NixOS/nixpkgs/nixos-23.05"},

	// ── JDK ───────────────────────────────────────────────────────────────
	"jdk8": {Registry: "github:NixOS/nixpkgs/nixos-23.11"},
}
