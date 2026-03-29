package config

type ProjectConfig struct {
	Name string `yaml:"name"`
	Version string `yaml:"version"`
	Dependencies Dependencies `yaml:"dependencies"`
	Hooks LifecycleHooks `yaml:"hooks"`
	Validations []ValidationCheck `yaml:"validations"`
}

type Dependencies struct {
	NixPackages []string `yaml:"nix_packages"`
	Dockerfile string `yaml:"docker_file,omitempty"`
}

type LifecycleHooks struct {
	PreInit string `yaml:"pre_init,omitempty"`
	PostInit string `yaml:"post_init,omitempty"`
	PreStart string `yaml:"pre_start,omitempty"`
	PostStart string `yaml:"post_start,omitempty"`
	PreBuild string `yaml:"pre_build,omitempty"`
	PostStop string `yaml:"post_stop,omitempty"`	
}

type ValidationCheck struct {
	Name string `yaml:"name"`
	Command string `yaml:"command"`
	AutoFix string `yaml:"auto_fix,omitempty"`
}