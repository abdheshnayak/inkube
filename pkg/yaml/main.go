package yaml

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/abdheshnayak/inkube/pkg/fn"
	yml "gopkg.in/yaml.v2"
)

type Config[T any] interface {
	Read() (*T, error)
	Write() error
}

type config[T any] struct {
	data *T
	path string
}

func (c *config[T]) Read() (*T, error) {
	t, err := ReadConfig[T](c.path)
	if err != nil {
		return c.data, err
	}

	c.data = t
	return t, nil
}

func (c *config[T]) Write() error {
	return WriteConfig(c.path, c.data, 0o644)
}

func GetHandler[T any](path string) Config[T] {
	data := new(T)
	return &config[T]{
		path: path,
		data: data,
	}
}

func ReadConfig[T any](path string) (*T, error) {
	var v T
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yml.Unmarshal(b, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func WriteConfig(path string, v any, perm fs.FileMode) error {

	b, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, create it
		if os.IsNotExist(err) {
			b, err = yml.Marshal(v)
			if err != nil {
				return fn.NewE(err)
			}

			return os.WriteFile(path, b, perm)
		}

		return fn.NewE(err)
	}

	var config yml.MapSlice
	if err := yml.Unmarshal(b, &config); err != nil {
		return fn.NewE(err)
	}

	if err := fillConfig(&config, v); err != nil {
		return fn.NewE(err)
	}

	b, err = yml.Marshal(config)
	if err != nil {
		return fn.NewE(err)
	}

	return os.WriteFile(path, b, perm)
}

// FillConfig updates src with values from dest
func fillConfig(src *yml.MapSlice, dest interface{}) error {
	// Marshal dest to YAML
	destBytes, err := yml.Marshal(dest)
	if err != nil {
		return fn.NewE(err)
	}

	// Unmarshal dest YAML into a MapSlice
	var destMapSlice yml.MapSlice
	if err := yml.Unmarshal(destBytes, &destMapSlice); err != nil {
		return fn.NewE(err)
	}

	// Iterate over destMapSlice and update src
	for _, item := range destMapSlice {
		key := item.Key
		value := item.Value

		// Check if key already exists in src, and update it if it does
		found := false
		for i, srcItem := range *src {
			if srcItem.Key == key {
				(*src)[i] = yml.MapItem{Key: key, Value: value}
				found = true
				break
			}
		}

		// If key doesn't exist in src, add it
		if !found {
			*src = append(*src, yml.MapItem{Key: key, Value: value})
		}
	}

	return nil
}
