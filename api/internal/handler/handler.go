package handler

import (
	"context"

	"github.com/rs/zerolog/log"
)

type handler struct {
}

func New() *handler {
	return &handler{}
}

func (h *handler) HealthCheck(ctx context.Context) (string, error) {
	logger := log.Ctx(ctx).With().Str("handler", "health-check").Logger()
	logger.Info().Msg("Health check")
	return "OK", nil
}
