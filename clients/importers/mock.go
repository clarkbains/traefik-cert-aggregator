package importers

import (
	"context"
	"crypto/x509"
	"errors"
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

func (v *MockClient) Start(ctx context.Context) error {
	var b = false
runLoop:
	for {
		b = !b
		v.manager.BeginChanges()
		if b {
			cert := x509.Certificate{SubjectKeyId: []byte{100, 68, 79, 63, 140}}
			v.manager.AddCert(&cert, nil, nil)
		}

		v.manager.DeleteUntouchedCerts()
		v.manager.EndChanges()
		select {
		case <-time.After(time.Second * 4):
		case <-ctx.Done():
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
		Name: "Mock",
		ConfigHelp: map[string]string{
			"foobar": "myhelp",
			"baz":    "foo",
		},
	}
}
