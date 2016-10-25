package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	gokitlog "github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

func newBackend(port int, logger gokitlog.Logger) error {
	svc := stockService("stock")

	var stockEndpoint endpoint.Endpoint
	stockEndpoint = newStockEndpoint(svc)
	stockEndpoint = newLoggingMw(gokitlog.NewContext(logger).With("layer", "endpoint"))(stockEndpoint)

	stockHandler := httptransport.NewServer(
		context.Background(),
		stockEndpoint,
		decodeRequest,
		encodeXMLResponse,
	)

	http.Handle("/", stockHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func newStockEndpoint(svc StockService) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return svc.Get(ctx)
	}
}
