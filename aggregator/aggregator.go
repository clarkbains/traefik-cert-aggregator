package aggregator

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"log"

	"sync"
)

var certUpdates = make(chan CertStoreChange, 10)
var certUpdatesOutgoing []chan CertStoreChange
var certUpdatesOutgoingLock sync.Mutex

func StartAggregating(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		var cm CertStoreChange

	runLoop:
		for {
			select {
			case cm = <-certUpdates:
				certUpdatesOutgoingLock.Lock()
				for _, channel := range certUpdatesOutgoing {
					channel <- cm
				}
				certUpdatesOutgoingLock.Unlock()
				log.Printf("\"%s\" has produced an update", cm.Sender)
			case <-ctx.Done():
				break runLoop
			}
		}
		certUpdatesOutgoingLock.Lock()
		defer certUpdatesOutgoingLock.Unlock()
		log.Println("Cleaning up exporter channels")
		for _, channel := range certUpdatesOutgoing {
			close(channel)
		}
	}()
	wg.Wait()
	log.Println("Core Aggregator finished")
}

type CertPackage struct {
	Cert  *x509.Certificate
	Chain []*x509.Certificate
	Key   *rsa.PrivateKey
}

type CertEntry struct {
	accessed bool
	certs    CertPackage
}

type CertManager struct {
	certs      map[string]CertEntry
	lock       sync.Mutex
	changeChan chan CertDiff
	name       string
	diff       CertDiff
}

type CertDiff struct {
	Added   []CertPackage
	Removed []CertPackage
}

type CertStoreChange struct {
	CertDiff CertDiff
	Sender   string
}

func NewOutputChan() chan CertStoreChange {
	certUpdatesOutgoingLock.Lock()
	defer certUpdatesOutgoingLock.Unlock()
	c := make(chan CertStoreChange, 5)
	certUpdatesOutgoing = append(certUpdatesOutgoing, c)
	return c
}

func NewCertManager(name string) *CertManager {
	cc := make(chan CertDiff)
	c := CertManager{
		changeChan: cc,
		name:       name,
		certs:      make(map[string]CertEntry),
	}

	go func() {
		for certDiff := range cc {
			csc := CertStoreChange{
				Sender:   c.name,
				CertDiff: certDiff,
			}
			certUpdates <- csc
		}
	}()

	return &c
}

func (c *CertManager) BeginChanges() {
	c.lock.Lock()
	for key, val := range c.certs {
		val.accessed = false
		c.certs[key] = val
	}

	c.diff.Added = c.diff.Added[:0]
	c.diff.Removed = c.diff.Removed[:0]

}

func (c *CertManager) EndChanges() {
	if len(c.diff.Added) > 0 || len(c.diff.Removed) > 0 {
		c.changeChan <- c.diff
	}

	c.lock.Unlock()
}

func (c *CertManager) GetDiff() CertDiff {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.diff
}

func (c *CertManager) DeleteUntouchedCerts() {
	for key, ce := range c.certs {
		if !ce.accessed {
			//log.Printf("Cleaning up cert %s", key)
			c.diff.Removed = append(c.diff.Removed, ce.certs)
			delete(c.certs, key)
		}
	}
}

func (c *CertManager) AddCert(cert *x509.Certificate, chain []*x509.Certificate, key *rsa.PrivateKey) bool {
	cip := c.IsCertInPool(cert)

	if !cip {
		//	log.Printf("Added new cert %s", string(cert.SubjectKeyId))
		ce := CertEntry{
			accessed: true,
			certs: CertPackage{
				Cert:  cert,
				Chain: chain,
				Key:   key,
			},
		}
		c.certs[(*cert.SerialNumber).String()] = ce
		c.diff.Added = append(c.diff.Added, ce.certs)
	}

	return !cip
}

func (c *CertManager) IsCertInPool(cert *x509.Certificate) bool {
	v, ok := c.certs[(*cert.SerialNumber).String()]
	if ok && v.certs.Cert.Equal(cert) {
		v.accessed = true
		c.certs[(*cert.SerialNumber).String()] = v
		return true
	}
	return false
}
