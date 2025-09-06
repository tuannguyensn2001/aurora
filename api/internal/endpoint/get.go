package endpoint

import (
	"context"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/rs/zerolog"
)

type IHandler interface {
	HealthCheck(ctx context.Context) (string, error)
}

type endpointList struct {
	HealthCheckHandler *httptransport.Server
}

func Get(svc IHandler, rootLogger zerolog.Logger) endpointList {

	healthCheckEndpoint := loggingMiddleware(rootLogger)(makeHealthCheckEndpoint(svc))

	return endpointList{
		HealthCheckHandler: httptransport.NewServer(
			healthCheckEndpoint,
			decodeHealthCheckRequest,
			encodeResponse,
		),
	}

}
