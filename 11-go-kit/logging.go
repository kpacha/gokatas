package main

import (
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

func newLoggingMw(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			logger.Log("action", "start", "request", request)
			begin := time.Now()

			response, err := next(ctx, request)

			logger.Log("action", "end", "response", response, "error", err, "took", time.Since(begin))

			return response, err
		}
	}
}
