package config

import (
	clientConfig "traefik-cert-aggregator/clients/config"
)

type Config struct {
	EnabledImporters []string
	EnabledExporters []string
	ImporterConfig   KeyedKVMap
	ExporterConfig   KeyedKVMap
}

type KeyedKVMap map[string](clientConfig.ClientConfiguration)

func Load() Config {
	cfg := Config{
		ImporterConfig: KeyedKVMap{},
		ExporterConfig: KeyedKVMap{},
	}
	return cfg
}

func (k *KeyedKVMap) Get(key string) clientConfig.ClientConfiguration {
	if _, ok := (*k)[key]; !ok {
		(*k)[key] = make(clientConfig.ClientConfiguration)
	}
	return (*k)[key]
}
