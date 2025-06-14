package config

import (
	"os"
	"path"
	"sync"

	cfhandler "github.com/abdheshnayak/inkube/pkg/config-handler"
)

type ConfigClient struct {
	*Config
	handler cfhandler.Config[Config]
	path    string
	mu      sync.RWMutex
}

func (c *ConfigClient) Write() error {
	return c.write()
}

func NewConfig() (*ConfigClient, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cpath := path.Join(cwd, "inkube.yaml")

	c := cfhandler.GetHandler[Config](cpath)
	resp := &ConfigClient{
		handler: c,
		Config:  &Config{},
		path:    cpath,
	}

	if err := resp.read(); err != nil {
		return nil, err
	}

	return resp, nil
}

var (
	config        *ConfigClient
	singletonOnce sync.Once = sync.Once{}
)

func Singleton() *ConfigClient {
	singletonOnce.Do(func() {
		var err error
		config, err = NewConfig()
		if err != nil {
			panic(err)
		}
		config.Write()
	})
	return config
}
