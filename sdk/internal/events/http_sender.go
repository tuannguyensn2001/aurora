package events

import (
	"context"
	"fmt"
	"sdk/pkg/errors"
	"sdk/pkg/logger"
	"sdk/types"

	"resty.dev/v3"
)

// HTTPEventSender implements EventSender using HTTP
type HTTPEventSender struct {
	endpointURL string
	logger      logger.Logger
}

// NewHTTPEventSender creates a new HTTP event sender
func NewHTTPEventSender(endpointURL string, logger logger.Logger) EventSender {
	return &HTTPEventSender{
		endpointURL: endpointURL,
		logger:      logger,
	}
}

// SendEvents sends events via HTTP
func (s *HTTPEventSender) SendEvents(ctx context.Context, events []types.EvaluationEvent) error {
	if len(events) == 0 {
		return nil
	}

	client := resty.New()
	defer client.Close()

	response, err := client.R().
		SetContext(ctx).
		SetBody(events).
		Post(fmt.Sprintf("%s/api/v1/sdk/events", s.endpointURL))

	if err != nil {
		return errors.NewNetworkError("send events", err)
	}

	if response.StatusCode() >= 400 {
		return errors.NewNetworkError("send events", fmt.Errorf("HTTP %d: %s", response.StatusCode(), response.String()))
	}

	s.logger.Debug("events sent successfully", "count", len(events))
	return nil
}
