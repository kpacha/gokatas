package main

type Pipes struct {
	Done chan struct{}
	Result chan *DataFormat
}

type Proxy struct {
	LoadBalancer *Balancer
}

func (p *Proxy) AddBackend(worker string, cancel <-chan struct{}) {
	p.LoadBalancer.AddBackend(worker, cancel)
}

func (p *Proxy) RemoveBackend(worker string, cancel <-chan struct{}) {
	p.LoadBalancer.RemoveBackend(worker, cancel)
}

func (p *Proxy) ProcessFirstResponse(pipes *Pipes) {
	responses := p.LoadBalancer.GetDataFromBackends(pipes.Done)
	for {
		select {
		case body := <-responses:
			go Parse(body, pipes)
		case <-pipes.Done:
			return
		}
	}
}