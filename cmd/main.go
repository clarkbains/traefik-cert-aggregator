package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
	"traefik-cert-aggregator/aggregator"
	"traefik-cert-aggregator/clients"
	"traefik-cert-aggregator/clients/exporters"
	"traefik-cert-aggregator/clients/importers"
	"traefik-cert-aggregator/config"
)
func getEnv(name string) string {
	return os.Getenv(name)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	cfg := config.Load()
	cfg.EnabledImporters = []string{"vault"}
	cfg.EnabledExporters = []string{"stdout", "traefik"}
	cfg.ImporterConfig = config.KeyedKVMap{
		"vault": map[string]string{
			"token": getEnv("VAULT_TOKEN"),
			"addr":  getEnv("VAULT_ADDR"),
		},
	}
	cfg.ExporterConfig = config.KeyedKVMap{
		"stdout": map[string]string{
			"prefix": "Stat exporter: ",
		},
		"traefik": map[string]string{
			"baseLocation": getEnv("TRAEFIK_BASE"),
		},
	}

	importers.AddAllClients(cfg)
	exporters.AddAllClients(cfg)
	wg.Add(2)
	go func() {
		defer wg.Done()

		err := clients.StartClients(&ctx, &cfg)
		if err != nil {
			log.Printf("Clients failed to configure: %s", err)
			cancel()

		}
	}()

	go func() {
		aggregator.StartAggregating(ctx)
		wg.Done()
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
		log.Printf("Got interupt. Cancelling executing goroutines")
		cancel()
		go func() {
			<- time.After(5*time.Second)
			panic("")
		}()
	case <-ctx.Done():
	}

	wg.Wait()
	log.Println("Bye!")

}
