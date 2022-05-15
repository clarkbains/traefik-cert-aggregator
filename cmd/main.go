package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"traefik-cert-aggregator/aggregator"
	"traefik-cert-aggregator/clients"
	"traefik-cert-aggregator/clients/exporters"
	"traefik-cert-aggregator/clients/importers"
	"traefik-cert-aggregator/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	cfg := config.Load()
	cfg.EnabledImporters = []string{"mock"}
	cfg.EnabledExporters = []string{"stdout"}
	cfg.ExporterConfig = config.KeyedKVMap{
		"stdout": map[string]string{
			"prefix":"Stat exporter: ",
		},
	}

	importers.AddAllClients(cfg)
	exporters.AddAllClients(cfg)
	wg.Add(2)
	go func() {
		defer wg.Done()

		err := clients.StartClients(ctx, &cfg)
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
	case <- c:
		log.Printf("Got interupt. Cancelling executing goroutines")
		cancel()
	case <- ctx.Done():
	}

	wg.Wait()
	log.Println("Bye!")

}
