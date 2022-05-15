package util

import "sync"

func WaitGroupToChannel(wg *sync.WaitGroup) chan struct{} {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		c <- struct{}{}
	}()
	return c
}
