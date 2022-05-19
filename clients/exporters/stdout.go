package exporters

import (
	"context"
	"errors"
	"log"
	"traefik-cert-aggregator/aggregator"
	"traefik-cert-aggregator/clients/config"
)

type StdoutExportClient struct {
	config config.ClientConfiguration
}

func NewStdoutExportClient() *StdoutExportClient {
	v := StdoutExportClient{}
	return &v
}

func (v *StdoutExportClient) Start(ctx *context.Context, ch chan aggregator.CertStoreChange) error {
	for {
		var cd aggregator.CertStoreChange
		var prefix = ""
		if val, ok := v.config["prefix"]; ok {
			prefix = val
		}
		select {
		case cd = <-ch:
			log.Printf("%sGot %d additions, %d removals from store \"%s\"", prefix, len(cd.CertDiff.Added), len(cd.CertDiff.Removed), cd.Sender)
		case <-(*ctx).Done():
			return errors.New("context cancelled")
		}
	}
}

func (v *StdoutExportClient) Configure(cc config.ClientConfiguration) error {
	v.config = cc
	return nil
}

func (v *StdoutExportClient) GetInfo() config.ClientInfo {
	return config.ClientInfo{
		Name: "stdout",
		ConfigHelp: map[string]string{
			"prefix": "prefix every message",
			"baz":    "foo",
		},
	}
}
