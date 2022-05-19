package config

import (
	"fmt"
)

type ClientConfiguration map[string]string

type ClientInfo struct {
	Name       string
	ConfigHelp map[string]string
}

func (c *ClientConfiguration) Get(key string, def string) string {
	if val, ok := (*c)[key]; ok {
		return val
	}

	return def
}

func (c *ClientConfiguration) GetErr(key string) (string, error) {
	if val, ok := (*c)[key]; ok {
		return val, nil
	}

	return "", fmt.Errorf("cannot get key \"%s\"", key)
}
