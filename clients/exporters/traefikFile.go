package exporters

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"traefik-cert-aggregator/aggregator"
	"traefik-cert-aggregator/clients/config"

	"gopkg.in/yaml.v3"
)

type TraefikConfig struct {
	Tls TraefikTlsConfig `yaml:"tls"`
}

type TraefikTlsConfig struct {
	Certificates []TraefikCertificateConfig `yaml:"certificates"`
}

type TraefikCertificateConfig struct {
	KeyFile  string `yaml:"keyFile"`
	CertFile string `yaml:"certFile"`
}
type TraefikExportClient struct {
	config        config.ClientConfiguration
	traefikConfig TraefikConfig
}

func NewTraefikExportClient() *TraefikExportClient {
	v := TraefikExportClient{}
	return &v
}

func (v *TraefikExportClient) Start(ctx *context.Context, ch chan aggregator.CertStoreChange) error {
	basePath := path.Clean(v.config.Get("baseLocation", os.TempDir()))

	for {
		var cd aggregator.CertStoreChange
		select {
		case cd = <-ch:
			for _, elem := range cd.CertDiff.Added {
				newPath := path.Join(basePath, cd.Sender, elem.Cert.SerialNumber.String())
				os.MkdirAll(newPath, 0711)
				keyPath := path.Join(newPath, "key.pem")
				log.Printf("Writing key and cert to %s (%s)", newPath, elem.Cert.Subject.CommonName)
				certPath := path.Join(newPath, "cert.pem")
				keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(elem.Key)})
				ioutil.WriteFile(keyPath, keyPEM, 0600)
				certStrings := make([]string, len(elem.Chain))
				for i, cert := range elem.Chain {
					certStrings[i] = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))
				}
				strings.Join(certStrings, "\n")
				ioutil.WriteFile(certPath, []byte(strings.Join(certStrings, "\n")), 0600)
			}

			for _, elem := range cd.CertDiff.Removed {
				newPath := path.Join(basePath, cd.Sender, elem.Cert.SerialNumber.String())
				err := os.RemoveAll(newPath)
				if err != nil {
					log.Printf("Could not remove stale certificate: %s", err)
					continue
				}
			}

			v.traefikConfig.Tls.Certificates = v.traefikConfig.Tls.Certificates[0:]

			importerFileinfo, err := ioutil.ReadDir(basePath)
			if err != nil {
				log.Printf("could not list directory to generate config: %s", err)
				continue
			}

			for _, importerDir := range importerFileinfo {
				if (!importerDir.IsDir()){
					continue
				}
				
				importerPath := path.Join(basePath, importerDir.Name())
				fileinfo, err := ioutil.ReadDir(importerPath)

				if err != nil {
					log.Printf("could not list directory to generate config: %s", err)
					continue
				}
				for _, file := range fileinfo {
					if !file.IsDir() {
						continue
					}
					keyPairPath := path.Join(importerPath, file.Name())
					keyPath, _ := filepath.Abs(path.Join(keyPairPath, "key.pem"))
					certPath, _ := filepath.Abs(path.Join(keyPairPath, "cert.pem"))
					_, err1 := os.Stat(keyPath)
					_, err2 := os.Stat(certPath)
					if err1 != nil || err2 != nil {
						log.Printf("Could not find key or cert file in dir: %s", keyPairPath)
						continue
					}
					
					tcc := TraefikCertificateConfig{
						KeyFile:  keyPath,
						CertFile: certPath,
					}

					v.traefikConfig.Tls.Certificates = append(v.traefikConfig.Tls.Certificates, tcc)
				}
			}

			traefikConfigFilePath := path.Join(basePath, "traefik.yaml")
			traefikCfgBytes, err := yaml.Marshal(v.traefikConfig)
			if err != nil {
				log.Printf("Could not create traefik config: %s", err)
			}
			ioutil.WriteFile(traefikConfigFilePath, traefikCfgBytes, 0600)

		case <-(*ctx).Done():
			return errors.New("context cancelled")
		}
	}
}

func (v *TraefikExportClient) Configure(cc config.ClientConfiguration) error {
	v.config = cc
	return nil
}

func (v *TraefikExportClient) GetInfo() config.ClientInfo {
	return config.ClientInfo{
		Name: "traefik",
		ConfigHelp: map[string]string{
			"baseLocation": "where to store all certs",
			"baz":          "foo",
		},
	}
}
