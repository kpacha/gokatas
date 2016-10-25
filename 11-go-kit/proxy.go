package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/context"

	jujuratelimit "github.com/juju/ratelimit"
	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	gokitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
)

type proxySettings struct {
	port         int
	retryMax     int
	retryTimeout time.Duration
	qps          int
	concurrency  int
}

func newProxy(settings proxySettings, client consulsd.Client, logger gokitlog.Logger) error {
	balancer := lb.NewRoundRobin(consulsd.NewSubscriber(
		client,
		backendsvcFactory("GET", "/"),
		gokitlog.NewContext(logger).With("layer", "consul subscriber"),
		BackendSVC,
		[]string{},
		true,
	))

	stockEndpoint := endpoint.Chain(
		newLoggingMw(gokitlog.NewContext(logger).With("layer", "http")),
		newRateLimitMw(settings.qps),
		newCircuitBreakerMw(gokitlog.NewContext(logger).With("layer", "circuitbreaker")),
		newConcurrentProxyMw(settings.concurrency),
	)(lb.Retry(settings.retryMax, settings.retryTimeout, balancer))

	stockHandler := httptransport.NewServer(
		context.Background(),
		stockEndpoint,
		decodeRequest,
		encodeJSONResponse,
	)

	http.Handle("/", stockHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", settings.port), nil)
}

func newRateLimitMw(qps int) endpoint.Middleware {
	return ratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(float64(qps), int64(qps)))
}

func newCircuitBreakerMw(logger gokitlog.Logger) endpoint.Middleware {
	settings := gobreaker.Settings{
		Interval: time.Duration(1) * time.Second,
		Timeout:  time.Duration(30) * time.Second,
	}
	return circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(settings))
}

func backendsvcFactory(method, path string) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}
		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		tgt.Path = path

		return httptransport.NewClient(method, tgt, encodeRequest, decodeXMLResponse).Endpoint(), nil, nil
	}
}

func newConcurrentProxyMw(total int) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			localCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			stock := make(chan interface{}, 1)
			errCh := make(chan error, total)
			for i := 0; i < total; i++ {
				go func() {
					response, err := next(localCtx, request)
					if err != nil {
						select {
						case <-ctx.Done():
						case errCh <- err:
						}
						return
					}
					select {
					case <-ctx.Done():
					case stock <- response:
						cancel()
					}
				}()
			}

			var err error
			for i := 0; i < total; i++ {
				select {
				case <-ctx.Done():
					return nil, err
				case s := <-stock:
					return s, nil
				case err = <-errCh:
				}
			}
			return nil, err
		}
	}
}
