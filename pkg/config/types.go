package config

type ResourceType string

type Deployment struct {
	Name      string `yaml:"name"`
	Container string `yaml:"container"`
}

type Config struct {
	Namespace  string     `yaml:"namespace"`
	Deployment Deployment `yaml:"deployment"`
	Intercept  bool       `yaml:"intercept"`
}
