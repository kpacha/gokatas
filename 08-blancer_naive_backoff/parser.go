package main

import (
	"encoding/xml"
)

type DataFormat struct {
	ProductList []struct {
		Sku      string `xml:"sku" json:"sku"`
		Quantity int    `xml:"quantity" json:"quantity"`
	} `xml:"Product" json:"products"`
}

func Parse(body []byte, pipes *Pipes) {
	data := &DataFormat{}
	err := xml.Unmarshal(body, data)
	if nil == err {
		select {
		case pipes.Result <- data:
		case <-pipes.Done:
		}
	}
}