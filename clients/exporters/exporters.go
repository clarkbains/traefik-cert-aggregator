package exporters

import (
	"traefik-cert-aggregator/clients"
	"traefik-cert-aggregator/config"
)

func AddAllClients(cfg config.Config) {
	//Configuration is provided here.
	clients.AddExportClient(NewStdoutExportClient())
	clients.AddExportClient(NewTraefikExportClient())
}
