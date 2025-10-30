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

	// Convert SDK events to API format
	apiEvents := make([]map[string]interface{}, len(events))
	for i, event := range events {
		apiEvent := map[string]interface{}{
			"id":             event.ID,
			"serviceName":    event.ServiceName,
			"eventType":      string(event.EventType),
			"parameterName":  event.ParameterName,
			"source":         event.Source,
			"userAttributes": event.UserAttributes,
			"timestamp":      event.Timestamp,
		}

		if event.RolloutValue != nil {
			apiEvent["rolloutValue"] = *event.RolloutValue
		}
		if event.Error != nil {
			apiEvent["error"] = *event.Error
		}
		if event.ExperimentID != nil {
			apiEvent["experimentId"] = *event.ExperimentID
		}
		if event.ExperimentUUID != nil {
			apiEvent["experimentUuid"] = *event.ExperimentUUID
		}
		if event.VariantID != nil {
			apiEvent["variantId"] = *event.VariantID
		}
		if event.VariantName != nil {
			apiEvent["variantName"] = *event.VariantName
		}

		apiEvents[i] = apiEvent
	}

	// Wrap in batch format expected by API
	batchRequest := map[string]interface{}{
		"events": apiEvents,
	}

	s.logger.Debug("sending events batch", "count", len(events), "endpoint", fmt.Sprintf("%s/api/v1/sdk/events", s.endpointURL))

	response, err := client.R().
		SetContext(ctx).
		SetBody(batchRequest).
		Post(fmt.Sprintf("%s/api/v1/sdk/events", s.endpointURL))

	if err != nil {
		s.logger.Error("failed to send events", "error", err, "count", len(events))
		return errors.NewNetworkError("send events", err)
	}

	if response.StatusCode() >= 400 {
		s.logger.Error("events API returned error", "status", response.StatusCode(), "body", response.String(), "count", len(events))
		return errors.NewNetworkError("send events", fmt.Errorf("HTTP %d: %s", response.StatusCode(), response.String()))
	}

	s.logger.Debug("events sent successfully", "count", len(events))
	return nil
}
