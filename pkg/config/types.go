package config

type LoadEnv struct {
	Name      *string `yaml:"name,omitempty"`
	Container string  `yaml:"container"`
	Enabled   bool    `yaml:"enabled"`

	Overrides map[string]string `yaml:"overrides"`
}

type TeleConfig struct {
	Name string `yaml:"name"`

	Connect   bool `yaml:"connect"`
	Intercept bool `yaml:"intercept"`
}

type Config struct {
	Version   string `yaml:"version"`
	Namespace string `yaml:"namespace"`

	Tele TeleConfig `yaml:"telepresence"`

	Devbox  bool    `yaml:"devbox"`
	LoadEnv LoadEnv `yaml:"loadEnv"`
}

type ConfigLock struct {
	Pkg string `yaml:"pkg"`
}
