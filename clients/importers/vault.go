package importers

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
	"traefik-cert-aggregator/aggregator"
	"traefik-cert-aggregator/clients/config"

	"github.com/hashicorp/vault/api"
)

type VaultClient struct {
	config  config.ClientConfiguration
	manager *aggregator.CertManager
	vault   *api.Client
}

type TLSEntry struct {
	PrivateKey *rsa.PrivateKey
	Chain []*x509.Certificate
}

func NewVaultClient() *VaultClient {
	v := VaultClient{}
	v.manager = aggregator.NewCertManager(v.GetInfo().Name)
	return &v
}



func (v *VaultClient) Start(ctx *context.Context) error {
runLoop:
	for {
		v.manager.BeginChanges()
		secret, err := v.vault.Logical().ListWithContext(*ctx, "kv/metadata/infrastructure/le-certs")
		if err != nil {
			return err
		}
		allVaultKeys, ok := secret.Data["keys"]
		if !ok {
			return errors.New("could not list keys")
		}

		var wg sync.WaitGroup
		domainData := allVaultKeys.([]interface{}) 
		parsedChan := make(chan TLSEntry, len(domainData)+1)
		
		for _, vaultKeyBasename := range domainData {
			wg.Add(1)
			go func (vaultKey string) {
				certSecret, err := v.vault.Logical().ReadWithContext(*ctx, "kv/data/infrastructure/le-certs/"+vaultKey)
				if err != nil {
					return
				}

				rawKVData, ok := certSecret.Data["data"]
				if !ok || rawKVData == nil {
					return
				}
				kvDataInterfaceMap := rawKVData.(map[string]interface{})
	
				foundKey, okm := kvDataInterfaceMap["key"].(string)
				foundCertChain, okk := kvDataInterfaceMap["cert"].(string)
				if !okm || !okk {
					log.Printf("unable to get data for %s", vaultKey)
					return
				}

				asyncParse(parsedChan, vaultKey, foundKey, foundCertChain)
				wg.Done()
			} (vaultKeyBasename.(string))
		}
		
		wg.Wait()
		close(parsedChan)
		for entry := range parsedChan {
			added := v.manager.AddCert(entry.Chain[0], entry.Chain, entry.PrivateKey)
			if added {
				log.Printf("vault: New cert for %s", entry.Chain[0].Subject.CommonName)
			}
		}

		v.manager.DeleteUntouchedCerts()
		v.manager.EndChanges()
		select {
		case <-time.After(time.Second * 10):
		case <-(*ctx).Done():
			break runLoop
		}
	}
	return errors.New("context cancelled")
}

func (v *VaultClient) Configure(cc config.ClientConfiguration) error {
	v.config = cc

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	client, err := api.NewClient(&api.Config{
		Address:    v.config.Get("addr", "https://localhost:8500"),
		HttpClient: httpClient,
	})

	token, err := v.config.GetErr("token")
	if err != nil {
		return err
	}
	client.SetToken(token)

	v.vault = client

	return nil
}

func (v *VaultClient) GetInfo() config.ClientInfo {
	return config.ClientInfo{
		Name: "vault",
		ConfigHelp: map[string]string{
			"token": "",
			"baz":   "foo",
		},
	}
}

func asyncParse(results chan TLSEntry, domain string, key string, chain string ) {
	var der, rest = pem.Decode([]byte(chain))
	var fullChain []*x509.Certificate
	for der != nil {
		singleCert, err := x509.ParseCertificate(der.Bytes)
		if err != nil {
			return
		}
		fullChain = append(fullChain, singleCert)
		der, rest = pem.Decode(rest)
	}

	pemBlock, _ := pem.Decode([]byte(key))
	parsedKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		log.Printf("Could not parse private key for %s, because %s", domain, err)
		return
	}
	te := TLSEntry{PrivateKey: parsedKey, Chain: fullChain}
	results <- te
}