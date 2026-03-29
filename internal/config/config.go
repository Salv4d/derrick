package config

type ProjectConfig struct {
	Name string `yaml:"name" validate:"required,lowercase"`
	Version string `yaml:"version" validate:"required"`
	Dependencies Dependencies `yaml:"dependencies" validate:"required"`
	Hooks LifecycleHooks `yaml:"hooks"`
	Validations []ValidationCheck `yaml:"validations" validate:"dive"`
}

type Dependencies struct {
	NixPackages []string `yaml:"nix_packages" validate:"required,min=1"`
	Dockerfile string `yaml:"docker_file,omitempty" validate:"omitempty,filepath"`
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
	Name string `yaml:"name" validate:"required"`
	Command string `yaml:"command" validate:"required"`
	AutoFix string `yaml:"auto_fix,omitempty"`
}