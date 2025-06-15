package config

type Environment struct {
}

type Devbox struct {
	Enabled bool           `yaml:"enabled"`
	Config  map[string]any `yaml:"config"`
}

type Intercept struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"`
}

type LoadEnv struct {
	Name      *string `yaml:"name,omitempty"`
	Container string  `yaml:"container"`
	Enabled   bool    `yaml:"enabled"`
}

type Config struct {
	Version   string `yaml:"version"`
	Namespace string `yaml:"namespace"`

	Intercept Intercept `yaml:"intercept"`
	Devbox    Devbox    `yaml:"devbox"`

	LoadEnv   LoadEnv           `yaml:"loadEnv"`
	LocalEnvs map[string]string `yaml:"envs"`
}
