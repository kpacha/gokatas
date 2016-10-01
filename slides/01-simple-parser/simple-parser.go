package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/rand"
	"time"
)

type DataFormat struct {
	ProductList []struct {
		Sku      string `xml:"sku" json:"sku"`
		Quantity int    `xml:"quantity" json:"quantity"`
	} `xml:"Product" json:"products"`
}

func main() {
	rand.Seed(time.Now().Unix())
	done := make(chan []byte, 10)

	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8" ?>
<ProductList>
    <Product>
        <sku>ABC123</sku>
        <quantity>2</quantity>
    </Product>
    <Product>
        <sku>ABC124</sku>
        <quantity>20</quantity>
    </Product>
</ProductList>`)

	for i := 0; i < 10; i++ {
		go parse(xmlData, done, i)
	}
	result := <-done
	fmt.Printf("json:\n%s\n-------------\n", result)
}

func parse(xmlData []byte, done chan<- []byte, parser int) {
	data := &DataFormat{}
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

	err := xml.Unmarshal(xmlData, data)
	if nil != err {
		fmt.Println("Error unmarshalling from XML", err)
		return
	}
	fmt.Printf("parser %d: XML unmashalled!\n", parser)
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

	result, err := json.Marshal(data)
	if nil != err {
		fmt.Println("Error marshalling to JSON", err)
		return
	}
	fmt.Printf("parser %d: JSON mashalled!\n", parser)
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

	done <- result
	fmt.Printf("parser %d XML:\n%s\n-------------\n", parser, xmlData)
	fmt.Printf("parser %d first sku:\n%s\n-------------\n", parser, data.ProductList[0].Sku)
}
