package endpoint

import (
	"api/internal/app"
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func makeHealthCheckEndpoint(svc IHandler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		result, err := svc.HealthCheck(ctx)
		if err != nil {
			return nil, err
		}

		return app.Response{
			Message: "Success",
			Data:    result,
		}, nil
	}
}

func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}
