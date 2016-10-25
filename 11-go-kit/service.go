package main

import (
	"errors"
	"math/rand"
	"time"

	"golang.org/x/net/context"
)

type StockServiceMiddleware func(StockService) StockService

type StockService interface {
	Get(ctx context.Context) (Stock, error)
}

type Stock struct {
	ProductList []Product `xml:"Product" json:"products"`
}

type Product struct {
	Sku      string `xml:"sku" json:"sku"`
	Quantity int    `xml:"quantity" json:"quantity"`
}

type stockService string

var ErrRandomError = errors.New("No one expects a random error!")

func (s stockService) Get(_ context.Context) (Stock, error) {
	if 90 < rand.Intn(100) {
		return Stock{}, ErrRandomError
	}
	s.fakeLoad()
	return s.randomStock(), nil
}

func (s stockService) fakeLoad() {
	rnd := rand.Intn(100)
	if 20 > rnd {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	} else if 70 > rnd {
		time.Sleep(time.Duration(rand.Intn(50)+50) * time.Millisecond)
	} else if 95 > rnd {
		time.Sleep(time.Duration(rand.Intn(500)+200) * time.Millisecond)
	}
}

func (s stockService) randomStock() Stock {
	products := make([]Product, rand.Intn(20))
	for i := range products {
		products[i] = s.randomProduct()
	}
	return Stock{products}
}

func (s stockService) randomProduct() Product {
	return Product{s.randSeq(40), rand.Intn(100)}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-.")

func (s stockService) randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
