package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
	"github.com/davecheney/profile"
)

const globalTimeout = 300 * time.Millisecond

func main() {
	port := flag.Int("port", 8080, "port")
	backends := flag.Int("workers", 3, "number of workers")
	strategy := flag.String("strategy", "majority", "balancing strategy ['one', 'two', 'majority', 'all']")
	poolSize := flag.Int("pool", 10, "size of the pool of available worker sets")
	flag.Parse()

	cfg := profile.Config{
		CPUProfile: true,
		MemProfile: true,
		ProfilePath: ".",
	}
	p := profile.Start(&cfg)
	defer p.Stop()

	proxy := Proxy{NewBalancer(initListOfBckends(backends), strategy, poolSize)}

	a := gin.Default()
	a.GET("/", func(c *gin.Context) {
		pipes := &Pipes {
			Done:		make(chan struct{}),
			Result:		make(chan *DataFormat),
		}
		go proxy.ProcessFirstResponse(pipes)
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
