package importers

import (
	"traefik-cert-aggregator/clients"
	"traefik-cert-aggregator/config"
)

func AddAllClients(cfg config.Config) {
	clients.AddImportClient(NewMockClient())
	clients.AddImportClient(NewVaultClient())
}
