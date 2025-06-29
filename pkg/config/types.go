package config

type LoadEnv struct {
	Name      *string `yaml:"name,omitempty"`
	Container string  `yaml:"container"`
	Enabled   bool    `yaml:"enabled"`

	Overrides map[string]string `yaml:"overrides"`
}

type BridgeConfig struct {
	Name string `yaml:"name"`

	Intercept bool `yaml:"intercept"`
}

type Config struct {
	Version   string `yaml:"version"`
	Namespace string `yaml:"namespace"`
	Connect   bool   `yaml:"connect"`

	Bridge BridgeConfig `yaml:"bridge"`

	Devbox  bool    `yaml:"devbox"`
	LoadEnv LoadEnv `yaml:"loadEnv"`
}

type ConfigLock struct {
	Pkg string `yaml:"pkg"`
}
