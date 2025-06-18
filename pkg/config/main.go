package config

import (
	"fmt"
	"os"
	"path"
	"sync"

	cfhandler "github.com/abdheshnayak/inkube/pkg/config-handler"
	"github.com/abdheshnayak/inkube/pkg/fn"
)

type ConfigClient struct {
	*Config
	handler cfhandler.Config[Config]
	path    string
	mu      sync.RWMutex
}

func (c *ConfigClient) Write() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.handler.Write()
}
func (c *ConfigClient) Reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cfg, err := c.handler.Read()
	if err != nil {
		return err
	}
	c.Config = cfg
	return nil
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
		mu:      sync.RWMutex{},
	}

	resp.mu.RLock()
	defer resp.mu.RUnlock()

	cfg, err := c.Read()
	if err != nil {
		return nil, err
	}

	resp.Config = cfg
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
			if os.IsNotExist(err) {
				fn.PrintError(fmt.Errorf("config file not found, please run `inkube init` first"))
				os.Exit(1)
			}

			fn.PrintError(err)
			os.Exit(1)
		}
		config.Write()
	})
	return config
}
