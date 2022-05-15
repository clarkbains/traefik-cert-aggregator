package config

type ClientConfiguration map[string]string

type ClientInfo struct {
	Name       string
	ConfigHelp map[string]string
}
