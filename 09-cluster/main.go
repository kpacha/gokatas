package main

import (
	"flag"
	"fmt"
	"strings"
	"github.com/gin-gonic/gin"
	"time"
	"github.com/davecheney/profile"
)

const globalTimeout = 300 * time.Millisecond

func main() {
	port := flag.Int("port", 8080, "port")
	backends := flag.String("workers", "", "knonw workers (ex: 'localhost:8081,localhost:8082')")
	strategy := flag.String("strategy", "majority", "balancing strategy ['one', 'two', 'majority', 'all']")
	poolSize := flag.Int("pool", 3, "size of the pool of available worker sets")
	flag.Parse()

	cfg := profile.Config{
		CPUProfile: true,
		MemProfile: true,
		ProfilePath: ".",
	}
	p := profile.Start(&cfg)
	defer p.Stop()

	proxy := Proxy{NewBalancer(initListOfBckends(backends), strategy, poolSize)}

	server := gin.Default()
	server.GET("/", func(c *gin.Context) {
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

	go func(){
		admin := gin.Default()
		admin.POST("/worker/*endpoint", func(c *gin.Context) {
			worker := c.Param("endpoint")
			done := make(chan struct{})
			go proxy.AddBackend(fmt.Sprintf("http:/%s/", worker), done)

			select {
			case <-done:
				c.String(200, "")
			case <-time.After(globalTimeout):
				c.String(500, "")
				close(done)
			}

		})
		admin.Run(fmt.Sprintf(":%d", *port - 10))
	}()

	server.Run(fmt.Sprintf(":%d", *port))
}

func initListOfBckends(backends *string) []string {
	if "" == *backends {
		return make([]string, 0)
	}
	workers := strings.Split(*backends, ",")
	for i, k := range workers {
		workers[i] = fmt.Sprintf("http://%s/", k)
	}
	return workers
}
