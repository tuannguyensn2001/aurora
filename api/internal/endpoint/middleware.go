package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/rs/zerolog"
)

func loggingMiddleware(logger zerolog.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			ctx = logger.WithContext(ctx)
			return next(ctx, request)
		}
	}
}
