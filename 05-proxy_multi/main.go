package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)

type DataFormat struct {
	ProductList []struct {
		Sku      string `xml:"sku" json:"sku"`
		Quantity int    `xml:"quantity" json:"quantity"`
	} `xml:"Product" json:"products"`
}

const globalTimeout = 300 * time.Millisecond

func main() {
	port := flag.Int("port", 8080, "port")
	backends := flag.Int("workers", 3, "number of workers")
	flag.Parse()

	workers := make([]string, *backends)
	for i := range workers {
		workers[i] = fmt.Sprintf("http://127.0.0.1:%d/", 8081+i)
	}

	a := gin.Default()
	a.GET("/", func(c *gin.Context) {
		timeouted := make(chan bool)
		result := getDataFromBackends(timeouted, workers)

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

func getDataFromBackends(timeOut <-chan bool, workers []string) chan *DataFormat {
	done := make(chan struct{})
	responses := make([]<-chan []byte, len(workers))

	for i := range workers {
		result := getDataFromBackend(done, workers[i])
		responses[i] = result
	}

	result := make(chan *DataFormat)
	for i := range responses {
		go func(response <-chan []byte) {
			select {
			case body := <-response:
				r, err := parse(body)
				if nil == err {
					result <- r
					defer close(done)
				}
			case <-timeOut:
				defer close(done)
			case <-done:
			}
		}(responses[i])
	}

	return result
}

func getDataFromBackend(done <-chan struct{}, backend string) <-chan []byte {
	work := make(chan []byte)
	go func() {
		resp, err := http.Get(backend)
		defer resp.Body.Close()
		if err == nil {
			body, _ := ioutil.ReadAll(resp.Body)
			select {
			case work <- body:
			case <-done:
			}
		}
	}()
	return work
}

func parse(xmlData []byte) (*DataFormat, error) {
	data := &DataFormat{}
	err := xml.Unmarshal(xmlData, data)
	return data, err
}
