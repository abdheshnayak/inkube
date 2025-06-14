package config

func (c *ConfigClient) read() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cfg, err := c.handler.Read()
	if err != nil {
		return err
	}

	c.Config = cfg
	return nil
}

func (c *ConfigClient) write() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.handler.Write()
}
