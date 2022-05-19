package config

import (
	clientConfig "traefik-cert-aggregator/clients/config"
)

type Config struct {
	EnabledImporters []string `env:"ENABLED_IMPORTERS"`
	EnabledExporters []string `env:"ENABLED_EXPORTERS"`
	ImporterConfig   KeyedKVMap `env:"IMPORTER_CONFIG"`
	ExporterConfig   KeyedKVMap `env:"EXPORTER_CONFIG"`
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
