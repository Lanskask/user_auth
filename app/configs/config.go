package configs

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type AllConfig struct {
	ServiceConfig      ServiceConfig      `yaml:"service"`
	SessionStoreConfig SessionStoreConfig `yaml:"session_store"`
	DBConfig           DBConfig           `yaml:"db"`
}

func getConfigFromFile[T any](filename string) (*T, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading the file %q: %s", filename, err)
	}

	c := new(T)

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("error marshalling a config file %q: %v", filename, err)
	}

	return c, nil
}
