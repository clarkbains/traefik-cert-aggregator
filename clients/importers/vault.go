package importers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"net/http"
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

	vaultKeys:
		for _, vaultKey := range allVaultKeys.([]interface{}) {

			certSecret, err := v.vault.Logical().ReadWithContext(*ctx, "kv/data/infrastructure/le-certs/"+vaultKey.(string))
			if err != nil {
				return err
			}
			rawKVData, ok := certSecret.Data["data"]
			if !ok || rawKVData == nil {
				//	log.Printf("unable to get cert data for %s", vaultKey)
				continue
			}

			kvDataInterfaceMap := rawKVData.(map[string]interface{})

			foundKey, okm := kvDataInterfaceMap["key"].(string)
			foundCertChain, okk := kvDataInterfaceMap["cert"].(string)
			if !okm || !okk {
				log.Printf("unable to get data for %s", vaultKey)
				continue
			}

			var der, rest = pem.Decode([]byte(foundCertChain))
			var fullChain []*x509.Certificate
			for der != nil {
				singleCert, err := x509.ParseCertificate(der.Bytes)
				if err != nil {
					continue vaultKeys
				}
				fullChain = append(fullChain, singleCert)
				der, rest = pem.Decode(rest)
			}

			pemBlock, _ := pem.Decode([]byte(foundKey))
			parsedKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
			if err != nil {
				log.Printf("Could not parse private key for %s, because %s", vaultKey, err)
				continue
			}
			added := v.manager.AddCert(fullChain[0], fullChain, parsedKey)
			if added {
				log.Printf("vault: New cert for %s", vaultKey)
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
