package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	port := flag.Int("port", 8081, "the port of the service")
	ip := flag.String("bind", "127.0.0.1", "the ip of the service")
	consumer := flag.String("consumer", "localhost:8070", "consumer hostname and 'admin' port")
	flag.Parse()
	fmt.Printf("Starting the flaky backend at port [%d]\n", *port)

	ticker := time.NewTicker(5 * time.Second)
	go func(ticker *time.Ticker) {
		uri := fmt.Sprintf("http://%s/worker/%s:%d", *consumer, *ip, *port)
		body := url.Values{}
		for _ = range ticker.C {
			http.PostForm(uri, body)
		}
	}(ticker)

	a := gin.Default()
	a.GET("/", func(c *gin.Context) {
		fakeLoad()
		if 90 < rand.Int31n(100) {
			c.String(500, "Internal server error")
		} else {
			products := ""
			for total := rand.Int31n(20); total > 0; total-- {
				products += MakeProduct()
			}
			c.String(200, fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n<ProductList>%s\n</ProductList>", products))
		}
	})
	a.Run(fmt.Sprintf(":%d", *port))
}

func fakeLoad() {
	rnd := rand.Int31n(100)
	if 20 > rnd {
		time.Sleep(time.Duration(rand.Int31n(10)) * time.Millisecond)
	} else if 70 > rnd {
		time.Sleep(time.Duration(rand.Int31n(50)+50) * time.Millisecond)
	} else if 95 > rnd {
		time.Sleep(time.Duration(rand.Int31n(500)+200) * time.Millisecond)
	}
}

func MakeProduct() string {
	return fmt.Sprintf("\n\t<Product>\n\t\t<sku>%s</sku>\n\t\t<quantity>%d</quantity>\n\t</Product>", randSeq(40), rand.Int31n(100))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-.")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
