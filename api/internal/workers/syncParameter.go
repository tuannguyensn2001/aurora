package workers

import (
	"api/config"
	"api/internal/dto"
	"api/internal/model"
	"api/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sdk"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
)

type SyncParameterWorker struct {
	river.WorkerDefaults[dto.SyncParameterArgs]
	Repository repository.Repository
	Cfg        config.Config
	S3         *s3.Client
}

func (w *SyncParameterWorker) Work(ctx context.Context, job *river.Job[dto.SyncParameterArgs]) error {
	logger := log.Ctx(ctx).With().Str("worker", "sync-parameter").Logger()
	logger.Info().Msg("Syncing parameter")
	err := w.ProcessSyncParameter(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to sync parameter")
		return err
	}

	return nil
}

func (w *SyncParameterWorker) ProcessSyncParameter(ctx context.Context) error {
	logger := log.Ctx(ctx).With().Str("worker", "sync-parameter").Logger()
	logger.Info().Msg("Processing sync parameter")

	parameters, err := w.Repository.GetAllParameters(ctx, 0, 0)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get parameter by ID")
		return err
	}

	// Convert to SDK format
	sdkParameters := make([]sdk.Parameter, len(parameters))
	for i, parameter := range parameters {
		sdkParameter, err := w.mapParameterToSDK(parameter)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to map parameter to SDK")
			return err
		}
		sdkParameters[i] = sdkParameter
	}
	logger.Info().Int("parameters_count", len(sdkParameters)).Msg("Found parameters to sync")
	jsonParameters, err := json.Marshal(sdkParameters)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal parameters")
		return err
	}

	_, err = w.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(w.Cfg.S3.BucketName),
		Key:    aws.String("parameters.json"),
		Body:   bytes.NewReader(jsonParameters),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to put parameters to S3")
		return err
	}

	return nil
}

func (w *SyncParameterWorker) mapParameterToSDK(parameter *model.Parameter) (sdk.Parameter, error) {

	if parameter == nil {
		return sdk.Parameter{}, errors.New("parameter is nil")
	}

	// Convert default rollout value to string
	defaultRolloutValueStr, err := w.rolloutValueToString(parameter.DefaultRolloutValue)
	if err != nil {
		return sdk.Parameter{}, err
	}

	// Map rules
	sdkRules, err := w.mapParameterRulesToSDK(parameter.Rules)
	if err != nil {
		return sdk.Parameter{}, err
	}

	return sdk.Parameter{
		Name:                parameter.Name,
		DataType:            sdk.ParameterDataType(parameter.DataType),
		DefaultRolloutValue: defaultRolloutValueStr,
		Rules:               sdkRules,
	}, nil
}

// rolloutValueToString converts a RolloutValue to its string representation
func (w *SyncParameterWorker) rolloutValueToString(rolloutValue model.RolloutValue) (string, error) {
	if rolloutValue.Data == nil {
		return "", nil
	}

	// Convert to JSON string representation
	jsonBytes, err := json.Marshal(rolloutValue.Data)
	if err != nil {
		return "", err
	}

	// Remove quotes if it's a simple string value
	jsonStr := string(jsonBytes)
	if len(jsonStr) >= 2 && jsonStr[0] == '"' && jsonStr[len(jsonStr)-1] == '"' {
		return jsonStr[1 : len(jsonStr)-1], nil
	}

	return jsonStr, nil
}

// mapParameterRulesToSDK converts model parameter rules to SDK parameter rules
func (w *SyncParameterWorker) mapParameterRulesToSDK(rules []model.ParameterRule) ([]sdk.ParameterRule, error) {
	sdkRules := make([]sdk.ParameterRule, len(rules))

	for i, rule := range rules {
		// Convert rollout value to string
		rolloutValueStr, err := w.rolloutValueToString(rule.RolloutValue)
		if err != nil {
			return nil, err
		}

		// Handle segment ID conversion (pointer uint to int64)
		var segmentID int64
		if rule.SegmentID != nil {
			segmentID = int64(*rule.SegmentID)
		}

		// Handle match type conversion (pointer to value)
		var matchType sdk.ConditionMatchType
		if rule.MatchType != nil {
			matchType = sdk.ConditionMatchType(*rule.MatchType)
		}

		// Map conditions
		sdkConditions, err := w.mapParameterRuleConditionsToSDK(rule.Conditions)
		if err != nil {
			return nil, err
		}

		sdkRules[i] = sdk.ParameterRule{
			ID:           rule.ID,
			Name:         rule.Name,
			Type:         sdk.RuleType(rule.Type),
			RolloutValue: rolloutValueStr,
			SegmentID:    segmentID,
			MatchType:    matchType,
			Conditions:   sdkConditions,
		}
	}

	return sdkRules, nil
}

// mapParameterRuleConditionsToSDK converts model parameter rule conditions to SDK parameter rule conditions
func (w *SyncParameterWorker) mapParameterRuleConditionsToSDK(conditions []model.ParameterRuleCondition) ([]sdk.ParameterRuleCondition, error) {
	sdkConditions := make([]sdk.ParameterRuleCondition, len(conditions))

	for i, condition := range conditions {
		// Get attribute information if available
		var attributeName, attributeDataType string
		if condition.Attribute != nil {
			attributeName = condition.Attribute.Name
			attributeDataType = string(condition.Attribute.DataType)
		}

		sdkConditions[i] = sdk.ParameterRuleCondition{
			ID:                condition.ID,
			AttributeID:       condition.AttributeID,
			Operator:          sdk.ConditionOperator(condition.Operator),
			Value:             condition.Value,
			AttributeName:     attributeName,
			AttributeDataType: attributeDataType,
		}
	}

	return sdkConditions, nil
}
