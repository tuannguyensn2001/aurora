package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/rs/zerolog"
)

func loggingMiddleware(log zerolog.Logger) endpoint.Middleware {

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			ctx = log.With().Logger().WithContext(ctx)
			return next(ctx, request)
		}
	}
}
