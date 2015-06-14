package main

import (
	"encoding/xml"
	"fmt"
	"github.com/plimble/ace"
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

const globalTimeout = 250 * time.Millisecond

func main() {
	rand.Seed(time.Now().Unix())
	a := ace.Default()
	a.GET("/", func(c *ace.C) {
		result := getDataFromBackend("http://127.0.0.1:8081/")

		select {
		case r := <-result:
			fmt.Printf("I have something! %s\n", r)
			if data, err := parse(r); nil == err {
				c.JSON(200, data)
			} else {
				fmt.Printf("I got an error parsing the data: %s\n", err)
				c.String(500, fmt.Sprintf("%s", err))
			}
		case <-time.After(globalTimeout):
			fmt.Println("Timeout!")
			c.String(500, "Backend did not respond")
		}

	})
	a.Run(":8080")
}

func getDataFromBackend(backend string) <-chan []byte {
	work := make(chan []byte)
	go func() {
		resp, err := http.Get(backend)
		if err != nil {
			// handle error
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		work <- body
	}()
	return work
}

func parse(xmlData []byte) (*DataFormat, error) {
	data := &DataFormat{}
	err := xml.Unmarshal(xmlData, data)
	if nil != err {
		fmt.Println("Error unmarshalling from XML", err)
		return nil, err
	}

	return data, nil
}
