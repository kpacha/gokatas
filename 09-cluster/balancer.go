package main

import (
	"io/ioutil"
	"net/http"
)

type Balancer struct {
	Membership *Membership
}

func NewBalancer(workers []string, strategy *string, poolSize *int) *Balancer {
	return &Balancer { NewMembership(workers, strategy, poolSize) }
}

func (b *Balancer) AddBackend(worker string, cancel chan struct{}) {
	b.Membership.AddBackend(worker, cancel)
}

func (b *Balancer) RemoveBackend(worker string, cancel chan struct{}) {
	b.Membership.RemoveBackend(worker, cancel)
}

func (b *Balancer) GetDataFromBackends(done chan struct{}) chan []byte {
	responses := make(chan []byte)
	go func() {
		select{
		case backends := <-b.Membership.NextWorkerSet:
			for _, backend := range backends {
				go b.getDataFromBackend(responses, done, backend)
			}
		case <-done:
			return
		}
	}()
	return responses
}

func (b *Balancer) getDataFromBackend(responses chan<- []byte, done <-chan struct{}, backend string) {
	if resp, err := http.Get(backend); err == nil {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			select {
			case responses <- body:
			case <-done:
			}
		}
	} else {
		b.Membership.MonitorWorker <- backend
	}
}
