package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/rand"
	"time"
)

type Stock struct {
	ProductList []struct {
		Sku      string `xml:"sku" json:"sku"`
		Quantity int    `xml:"quantity" json:"quantity"`
	} `xml:"Product" json:"products"`
}

func main() {
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

	result, err := parse(xmlData)
	if nil != err {
		fmt.Println("Error parsing data", err.Error())
		panic(err)
	}
	fmt.Printf("json:\n%s\n-------------\n", result)
}

func parse(xmlData []byte) ([]byte, error) {
	data := &Stock{}
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

	err := xml.Unmarshal(xmlData, data)
	if nil != err {
		fmt.Println("Error unmarshalling from XML", err.Error())
		return []byte{}, err
	}
	fmt.Println("XML unmashalled!\n")
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

	result, err := json.Marshal(data)
	if nil != err {
		fmt.Println("Error marshalling to JSON", err.Error())
		return []byte{}, err
	}
	fmt.Println("JSON mashalled!\n")
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

	fmt.Println("first sku:\n%s\n-------------\n", data.ProductList[0].Sku)

	return result, nil
}
