package config

import (
	"os"
)

func (c *configImpl) read() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	b, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			c.Config = &Config{}

			c.write()
			return nil
		}
	}

	return nil
}

func (c *configImpl) write() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return os.WriteFile(c.path, nil, 0644)
}
