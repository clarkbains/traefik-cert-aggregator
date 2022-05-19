package importers

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"log"
	"math/big"
	"time"
	"traefik-cert-aggregator/aggregator"
	"traefik-cert-aggregator/clients/config"
)

type MockClient struct {
	config  config.ClientConfiguration
	manager *aggregator.CertManager
}

func NewMockClient() *MockClient {
	v := MockClient{}
	v.manager = aggregator.NewCertManager(v.GetInfo().Name)
	return &v
}

func (v *MockClient) Start(ctx *context.Context) error {
	log.Println("mock started")
	var b = false
runLoop:
	for {
		b = !b
		v.manager.BeginChanges()
		if b {
			cert := x509.Certificate{SerialNumber: big.NewInt(10567)}
			fc := []*x509.Certificate{&cert}
			pk := rsa.PrivateKey{}
			v.manager.AddCert(&cert, fc, &pk)
		}

		v.manager.DeleteUntouchedCerts()
		v.manager.EndChanges()
		select {
		case <-time.After(time.Second * 4):
		case <-(*ctx).Done():
			break runLoop
		}
	}
	return errors.New("context cancelled")
}

func (v *MockClient) Configure(cc config.ClientConfiguration) error {
	v.config = cc
	return nil
}

func (v *MockClient) GetInfo() config.ClientInfo {
	return config.ClientInfo{
		Name: "mock",
		ConfigHelp: map[string]string{
			"foobar": "myhelp",
			"baz":    "foo",
		},
	}
}
