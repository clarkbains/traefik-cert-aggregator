package aggregator

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"log"

	"sync"
)

var certUpdates = make(chan *CertManager, 10)
var certUpdatesOutgoing []chan CertStoreChange
var certUpdatesOutgoingLock sync.Mutex

func StartAggregating(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		var cm *CertManager

	runLoop:
		for {
			select {
			case cm = <-certUpdates:
				certUpdatesOutgoingLock.Lock()
				csc := CertStoreChange{
					CertDiff: cm.GetDiff(),
					Sender:   cm.name,
				}
				for _, channel := range certUpdatesOutgoing {
					channel <- csc
				}
				certUpdatesOutgoingLock.Unlock()
				log.Printf("\"%s\" has produced an update", cm.name)
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
	Chain *x509.CertPool
	Key   *rsa.PrivateKey
}

type CertEntry struct {
	accessed bool
	certs    CertPackage
}

type CertManager struct {
	certs      map[string]CertEntry
	lock       sync.Mutex
	changeChan chan struct{}
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
	cc := make(chan struct{})
	c := CertManager{
		changeChan: cc,
		name:       name,
		certs:      make(map[string]CertEntry),
	}

	go func() {
		for range cc {
			certUpdates <- &c
		}
	}()

	return &c
}

func (c *CertManager) BeginChanges() {
	c.lock.Lock()
	//log.Printf("Manager for \"%s\" has started collecting changes", c.name)
	for key, val := range c.certs {
		val.accessed = false
		c.certs[key] = val
	}

	c.diff.Added = c.diff.Added[:0]
	c.diff.Removed = c.diff.Removed[:0]

}

func (c *CertManager) EndChanges() {
	if len(c.diff.Added) > 0 || len(c.diff.Removed) > 0 {
		c.changeChan <- struct{}{}
	}
	//log.Printf("Manager for \"%s\" has finished collecting changes", c.name)

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
			//	log.Printf("Cleaning up cert %s", key)
			c.diff.Removed = append(c.diff.Removed, ce.certs)
			delete(c.certs, key)
		}
	}
}

func (c *CertManager) AddCert(cert *x509.Certificate, chain *x509.CertPool, key *rsa.PrivateKey) {
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
		c.certs[string(cert.SubjectKeyId)] = ce
		c.diff.Added = append(c.diff.Added, ce.certs)
	}
}

func (c *CertManager) IsCertInPool(cert *x509.Certificate) bool {
	v, ok := c.certs[string(cert.SubjectKeyId)]
	if ok && v.certs.Cert.Equal(cert) {
		log.Printf("Accessedold cert %s", string(cert.SubjectKeyId))

		v.accessed = true
		return true
	}
	return false
}
