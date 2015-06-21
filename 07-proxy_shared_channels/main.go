package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
	"math/rand"
	"github.com/davecheney/profile"
)

type DataFormat struct {
	ProductList []struct {
		Sku      string `xml:"sku" json:"sku"`
		Quantity int    `xml:"quantity" json:"quantity"`
	} `xml:"Product" json:"products"`
}

type Balancer struct {
	Workers []string
	TotalWorkers int
	MaxWorkersOnAJob int
	Strategy string
}

type Pipes struct {
	Done chan struct{}
	Result chan *DataFormat
}

const globalTimeout = 300 * time.Millisecond

func main() {
	port := flag.Int("port", 8080, "port")
	backends := flag.Int("workers", 3, "number of workers")
	strategy := flag.String("strategy", "majority", "balancing strategy ['one', 'two', 'majority', 'all']")
	flag.Parse()

	cfg := profile.Config{
		CPUProfile:		true,
		MemProfile:     true,
		ProfilePath:    ".",
	}
	p := profile.Start(&cfg)
	defer p.Stop()

	balancer := newBalancer(initListOfBckends(backends), strategy)

	a := gin.Default()
	a.GET("/", func(c *gin.Context) {
		pipes := &Pipes {
			Done:		make(chan struct{}),
			Result:		make(chan *DataFormat),
		}
		go processFirstResponse(pipes, balancer)
		defer close(pipes.Done)

		select {
		case data := <-pipes.Result:
			c.JSON(200, data)
		case <-time.After(globalTimeout):
			c.JSON(500, nil)
		}

	})
	a.Run(fmt.Sprintf(":%d", *port))

}

func initListOfBckends(totalWorkers *int) []string {
	workers := make([]string, *totalWorkers)
	for i := range workers {
		workers[i] = fmt.Sprintf("http://127.0.0.1:%d/", 8081+i)
	}
	return workers
}

func newBalancer(workers []string, strategy *string) *Balancer {
	fmt.Println("Creating a new Blancer for the backends:", workers)
	return &Balancer {
		Workers: workers,
		TotalWorkers: len(workers),
		MaxWorkersOnAJob: getMaxWorkersOnAJob(strategy, len(workers)),
		Strategy: *strategy,
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
			maxWorkersOnAJob = totalWorkers / 2 + 1
		case "all":
			maxWorkersOnAJob = totalWorkers
	}
	if maxWorkersOnAJob > totalWorkers {
		maxWorkersOnAJob = totalWorkers
	}
	return maxWorkersOnAJob
}

func processFirstResponse(pipes *Pipes, balancer *Balancer) {
	responses := balancer.GetDataFromBackends(pipes.Done)
	for {
		select {
		case body := <-responses:
			go parse(body, pipes)
		case <-pipes.Done:
			return
		}
	}
}

func parse(body []byte, pipes *Pipes) {
	data := &DataFormat{}
	err := xml.Unmarshal(body, data)
	if nil == err {
		select {
		case pipes.Result <- data:
		case <-pipes.Done:
		}
	}
}

func (b *Balancer) GetDataFromBackends(done chan struct{}) chan []byte {
	responses := make(chan []byte)
	if 0 > b.TotalWorkers {
		workerIds := rand.Perm(b.TotalWorkers)
		for i := 0; i < b.MaxWorkersOnAJob; i++ {
			go b.GetDataFromBackend(responses, done, &workerIds[i])
		}
	}
	return responses
}

func (b *Balancer) GetDataFromBackend(responses chan<- []byte, done <-chan struct{}, backendId *int) {
	if resp, err := http.Get(b.Workers[*backendId]); err == nil {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			select {
			case responses <- body:
			case <-done:
			}
		}
	}
}
