package config

type Config struct {
	Version string `yaml:"version"`

	Namespace string `yaml:"namespace"`
	Name      string `yaml:"name"`
	Container string `yaml:"container"`

	Intercept bool `yaml:"intercept"`
	Loadenv   bool `yaml:"loadenv"`
	Devbox    bool `yaml:"devbox"`
}
