package main

import (
	"math/rand"
	"net/http"
	"time"
)

type Membership struct {
	Workers          []string
	TotalWorkers     int
	MaxWorkersOnAJob int
	Strategy         string
	RemoveWorker     chan string
	AddWorker        chan string
	NextWorkerSet    chan []string
	MonitorWorker    chan string
	Backoff          BackoffPolicy
}

func NewMembership(workers []string, strategy *string, poolSize *int) *Membership {
	m := &Membership{
		Workers:          workers,
		TotalWorkers:     len(workers),
		Strategy:         *strategy,
		MaxWorkersOnAJob: getMaxWorkersOnAJob(strategy, len(workers)),
		RemoveWorker:     make(chan string),
		AddWorker:        make(chan string),
		NextWorkerSet:    make(chan []string, *poolSize),
		MonitorWorker:    make(chan string),
		Backoff:          Default,
	}
	for i := 0; i < *poolSize; i++ {
		m.NextWorkerSet <- m.newWorkerSet()
	}
	go m.manageWorkerPool()
	return m
}

func (m *Membership) AddBackend(worker string, cancel <-chan struct{}) {
	select {
	case m.AddWorker <- worker:
	case <-cancel:
	}
}

func (m *Membership) RemoveBackend(worker string, cancel <-chan struct{}) {
	select {
	case m.RemoveWorker <- worker:
	case <-cancel:
	}
}

func getMaxWorkersOnAJob(strategy *string, totalWorkers int) int {
	maxWorkersOnAJob := 0
	switch *strategy {
	case "one":
		maxWorkersOnAJob = 1
	case "two":
		maxWorkersOnAJob = 2
	case "majority":
		maxWorkersOnAJob = totalWorkers/2 + 1
	case "all":
		maxWorkersOnAJob = totalWorkers
	}
	if maxWorkersOnAJob > totalWorkers {
		maxWorkersOnAJob = totalWorkers
	}
	return maxWorkersOnAJob
}

func (m *Membership) manageWorkerPool() {
	for {
		select {
		case worker := <-m.MonitorWorker:
			if 0 <= m.remove(worker) {
				go m.rejoinWorker(worker, 1)
			}
		case worker := <-m.RemoveWorker:
			m.remove(worker)
		case worker := <-m.AddWorker:
			if workerId := m.getWorkerId(worker); -1 == workerId {
				m.rescale(append(m.Workers, worker))
			}
		case m.NextWorkerSet <- m.newWorkerSet():
		}
	}
}

func (m *Membership) remove(worker string) int {
	workerId := m.getWorkerId(worker)
	if 0 <= workerId {
		workers := append(m.Workers[:workerId], m.Workers[workerId+1:]...)
		m.rescale(workers)
	}
	return workerId
}

func (m *Membership) newWorkerSet() []string {
	workerSet := make([]string, m.MaxWorkersOnAJob)
	workerIds := rand.Perm(m.TotalWorkers)
	for i := range workerSet {
		workerSet[i] = m.Workers[workerIds[i]]
	}
	return workerSet
}

func (m *Membership) getWorkerId(backend string) int {
	for k, v := range m.Workers {
		if v == backend {
			return k
		}
	}
	return -1
}

func (m *Membership) rescale(workers []string) {
	m.MaxWorkersOnAJob = getMaxWorkersOnAJob(&(m.Strategy), len(workers))
	m.TotalWorkers = len(workers)
	m.Workers = workers
}

func (m *Membership) rejoinWorker(worker string, attempt int) {
	if m.Backoff.MaxRetries <= attempt {
		return
	}
	chekWorker := time.After(m.Backoff.Duration(attempt))
	select {
	case <-chekWorker:
		if resp, err := http.Get(worker); err == nil {
			defer resp.Body.Close()
			m.AddWorker <- worker
		} else {
			m.rejoinWorker(worker, attempt+1)
		}
	}
}
