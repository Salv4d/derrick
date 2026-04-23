package engine

import (
	"os"
	"os/exec"
	"strings"
)

// Chain handles the recursive execution tracking to prevent infinite loops (fork-bombs)
// in project dependency trees.
type Chain struct {
	EnvVar string
	Names  map[string]bool
	Raw    string
}

// GetChain loads the recursion chain from the environment.
func GetChain(envVar string) *Chain {
	raw := os.Getenv(envVar)
	names := make(map[string]bool)
	if raw != "" {
		for _, name := range strings.Split(raw, ",") {
			if name = strings.TrimSpace(name); name != "" {
				names[name] = true
			}
		}
	}
	return &Chain{
		EnvVar: envVar,
		Names:  names,
		Raw:    raw,
	}
}

// Contains returns true if the project is already in the recursion chain.
func (c *Chain) Contains(name string) bool {
	return c.Names[name]
}

// Next returns the environment string for the next level of recursion.
func (c *Chain) Next(currentProject string) string {
	if c.Raw == "" {
		return currentProject
	}
	return c.Raw + "," + currentProject
}

// ResolveDerrickBinary returns the absolute path to the current derrick binary.
func ResolveDerrickBinary() string {
	exe, err := os.Executable()
	if err != nil {
		return os.Args[0]
	}
	return exe
}

// ExecuteRecursive runs a derrick command in a child directory with the
// appropriate recursion environment.
func ExecuteRecursive(dir, command string, profile string, chain *Chain, currentProject string, extraArgs []string, extraEnv []string) error {
	binary := ResolveDerrickBinary()

	args := []string{command}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	args = append(args, extraArgs...)

	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	env := append(os.Environ(), chain.EnvVar+"="+chain.Next(currentProject))
	cmd.Env = append(env, extraEnv...)

	return cmd.Run()
}
