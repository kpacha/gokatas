package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/davecheney/profile"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type DataFormat struct {
	ProductList []struct {
		Sku      string `xml:"sku" json:"sku"`
		Quantity int    `xml:"quantity" json:"quantity"`
	} `xml:"Product" json:"products"`
}

type Balancer struct {
	Workers          []string
	TotalWorkers     int
	MaxWorkersOnAJob int
}

type Pipes struct {
	Done    chan struct{}
	Result  chan *DataFormat
	TimeOut <-chan bool
}

const globalTimeout = 300 * time.Millisecond

func main() {
	port := flag.Int("port", 8080, "port")
	backends := flag.Int("workers", 3, "number of workers")
	strategy := flag.String("strategy", "majority", "balancing strategy ['one', 'two', 'majority', 'all']")
	flag.Parse()

	cfg := profile.Config{
		CPUProfile:  true,
		MemProfile:  true,
		ProfilePath: ".",
	}
	p := profile.Start(&cfg)
	defer p.Stop()

	balancer := newBalancer(backends, strategy)

	a := gin.Default()
	a.GET("/", func(c *gin.Context) {
		timeouted := make(chan bool)
		result := processFirstResponse(timeouted, balancer)

		select {
		case data := <-result:
			c.JSON(200, data)
		case <-time.After(globalTimeout):
			c.JSON(500, nil)
			timeouted <- true
		}

	})
	a.Run(fmt.Sprintf(":%d", *port))

}

func newBalancer(totalWorkers *int, strategy *string) *Balancer {
	workers := make([]string, *totalWorkers)
	for i := range workers {
		workers[i] = fmt.Sprintf("http://127.0.0.1:%d/", 8081+i)
	}
	maxWorkersOnAJob := 0
	switch *strategy {
	case "one":
		maxWorkersOnAJob = 1
	case "two":
		maxWorkersOnAJob = 2
	case "majority":
		maxWorkersOnAJob = len(workers)/2 + 1
	case "all":
		maxWorkersOnAJob = len(workers)
	}
	if maxWorkersOnAJob > len(workers) {
		maxWorkersOnAJob = len(workers)
	}
	return &Balancer{
		Workers:          workers,
		TotalWorkers:     len(workers),
		MaxWorkersOnAJob: maxWorkersOnAJob,
	}
}

func processFirstResponse(timeOut <-chan bool, balancer *Balancer) chan *DataFormat {
	pipes := &Pipes{
		Done:    make(chan struct{}),
		Result:  make(chan *DataFormat),
		TimeOut: timeOut,
	}
	responses := balancer.GetDataFromBackends(pipes.Done)
	for _, resp := range responses {
		go parseResponse(resp, pipes)
	}

	return pipes.Result
}

func parseResponse(response <-chan []byte, pipes *Pipes) {
	select {
	case body := <-response:
		r, err := parse(body)
		if nil == err {
			pipes.Result <- r
			defer close(pipes.Done)
		}
	case <-pipes.TimeOut:
		defer close(pipes.Done)
	case <-pipes.Done:
	}
}

func parse(xmlData []byte) (*DataFormat, error) {
	data := &DataFormat{}
	err := xml.Unmarshal(xmlData, data)
	return data, err
}

func (b *Balancer) GetDataFromBackends(done chan struct{}) []<-chan []byte {
	responses := make([]<-chan []byte, b.MaxWorkersOnAJob)
	workerIds := rand.Perm(b.TotalWorkers)
	for i := range responses {
		responses[i] = b.GetDataFromBackend(done, &(b.Workers[workerIds[i]]))
	}
	return responses
}

func (b *Balancer) GetDataFromBackend(done <-chan struct{}, backend *string) <-chan []byte {
	work := make(chan []byte)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in GetDataFromBackend: ", *backend)
			}
		}()
		resp, err := http.Get(*backend)
		if err == nil {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			select {
			case work <- body:
			case <-done:
			}
		}
	}()
	return work
}
