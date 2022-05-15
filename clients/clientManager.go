package clients

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"
	"traefik-cert-aggregator/aggregator"
	clientConfig "traefik-cert-aggregator/clients/config"
	"traefik-cert-aggregator/config"
	"traefik-cert-aggregator/util"
)

type Client interface {
	Configure(clientConfig.ClientConfiguration) error
	GetInfo() clientConfig.ClientInfo
}

type ImportClient interface {
	Client
	Start(context.Context) error
}

type ExportClient interface {
	Client
	Start(context.Context, chan aggregator.CertStoreChange) error
}

var registeredImportClients []ImportClient
var registeredExportClients []ExportClient

func toClientSlice[T Client](slice []T) []Client {
	var ret []Client
	for _, v := range slice {
		ret = append(ret, v)
	}
	return ret
}

func StartClients(ctx context.Context, cfg *config.Config) error {
	wg := sync.WaitGroup{}

	ic := toClientSlice(registeredImportClients)
	importers := configureClients[ImportClient](&cfg.ImporterConfig, cfg.EnabledImporters, &ic)

	ec := toClientSlice(registeredExportClients)
	exporters := configureClients[ExportClient](&cfg.ExporterConfig, cfg.EnabledExporters, &ec)
	if len(exporters) == 0 || len(importers) == 0 {
		log.Printf("%d importer clients, %d exporter clients configured", len(importers), len(exporters))
		return errors.New("no client configured for running")
	}
	wg.Add(2)
	go func() {
		runImportClients(ctx, importers)
		wg.Done()
	}()

	go func() {
		runExportClients(ctx, exporters)
		wg.Done()
	}()

	wg.Wait()
	log.Println("All Clients Finished")
	return nil
}

func AddImportClient(c ImportClient) {
	registeredImportClients = append(registeredImportClients, c)
	log.Printf("Discovered import client \"%s\"", c.GetInfo().Name)
}

func AddExportClient(c ExportClient) {
	registeredExportClients = append(registeredExportClients, c)
	log.Printf("Discovered export client \"%s\"", c.GetInfo().Name)

}

func configureClients[T Client](cfg *config.KeyedKVMap, clientWhitelist []string, clientMap *[]Client) []T {
	var configured []T

	allowed := util.NewSet[string]()
	for _, itm := range clientWhitelist {
		allowed.Add(strings.ToLower(itm))
	}

	for _, client := range *clientMap {
		name :=  client.GetInfo().Name 
		if !allowed.Contains(strings.ToLower(client.GetInfo().Name) ) {
			log.Printf("Client \"%s\" not whitelisted", name)
			continue
		}
		log.Printf("Initiallizing client \"%s\"", name)
		err := client.Configure(cfg.Get(name))
		if err != nil {
			log.Fatalf("Error while configuring \"%s\": %s", name, err)
			continue
		}
		configured = append(configured, client.(T))
	}
	return configured
}

func runImportClients(ctx context.Context, clientSet []ImportClient) {
	wg := sync.WaitGroup{}
	for _, client := range clientSet {
		wg.Add(1)
		go func() {
		runLoop:
			for {
				err := client.Start(ctx)
				select {
				case <-time.After(time.Second):
					log.Printf("Import client \"%s\" terminated with the following error. Restarting. %s", client.GetInfo().Name, err)
				case <-ctx.Done():
					break runLoop
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func runExportClients(ctx context.Context, clientSet []ExportClient) {
	wg := sync.WaitGroup{}
	for _, client := range clientSet {
		wg.Add(1)
		go func() {

		runLoop:
			for {
				err := client.Start(ctx, aggregator.NewOutputChan())
				select {
				case <-time.After(time.Second):
					log.Printf("Export client \"%s\" terminated with the following error. Restarting. %s", client.GetInfo().Name, err)
				case <-ctx.Done():
					break runLoop
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
