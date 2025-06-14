package config

import (
	"os"
	"path"
	"sync"
)

type Config struct {
}

type ConfigClient interface {
	Get() *Config
	Write() error
}

type configImpl struct {
	*Config
	path string
	mu   sync.RWMutex
}

func (c *configImpl) Write() error {
	return c.write()
}

func (c *configImpl) Get() *Config {
	return c.Config
}

func NewConfig() (ConfigClient, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	resp := &configImpl{
		Config: &Config{},
		path:   path.Join(cwd, "inkube.yaml"),
	}

	if err := resp.read(); err != nil {
		return nil, err
	}

	return resp, nil
}

var (
	config        ConfigClient
	singletonOnce sync.Once
)

func Singleton() (ConfigClient, error) {
	singletonOnce.Do(func() {
		config, err := NewConfig()
		if err != nil {
			panic(err)
		}
		config.Write()
	})
	return config, nil
}
