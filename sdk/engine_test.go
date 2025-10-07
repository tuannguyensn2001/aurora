package sdk

import (
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEvaluateExperiment(t *testing.T) {
	t.Run("population size 100, no segment", func(t *testing.T) {
		now := time.Now()
		experiment := Experiment{
			ID:                1,
			Name:              "test experiment",
			Uuid:              "b7fb8970-3fe7-49a7-9b3a-f670a7ae7641",
			StartDate:         now.Add(-1 * time.Hour).Unix(),
			EndDate:           now.Add(1 * time.Hour).Unix(),
			HashAttributeID:   1,
			PopulationSize:    100,
			Strategy:          "percentage_split",
			HashAttributeName: "user_id",
			Status:            ExperimentStatusRunning,
			Variants: []ExperimentVariant{
				{
					ID:                1,
					Name:              "test variant",
					TrafficAllocation: 50,
					Parameters: []ExperimentVariantParameter{
						{
							ID:                1,
							ParameterDataType: ParameterDataTypeBoolean,
							ParameterID:       1,
							ParameterName:     "enableAuth",
							RolloutValue:      "false",
						},
					},
				},
				{
					ID:                2,
					Name:              "test variant 2",
					TrafficAllocation: 50,
					Parameters: []ExperimentVariantParameter{
						{
							ID:                2,
							ParameterDataType: ParameterDataTypeBoolean,
							ParameterID:       2,
							ParameterName:     "enableAuth",
							RolloutValue:      "true",
						},
					},
				},
			},
		}

		engine := newEngine(slog.New(slog.NewTextHandler(os.Stdout, nil)))
		control := 0
		treatment := 0
		total := 1000
		for i := 1; i <= total; i++ {
			userID := i
			attribute := NewAttribute().SetNumber("user_id", float64(userID))
			rolloutValue, dataType, ok := engine.evaluateExperiment(&experiment, attribute, "enableAuth")
			require.Equal(t, true, ok)
			require.Equal(t, ParameterDataTypeBoolean, dataType)
			if rolloutValue == "true" {
				treatment++
			} else if rolloutValue == "false" {
				control++
			}
		}
		log.Println(control, treatment)
		require.Equal(t, total, control+treatment)
	})

	t.Run("population size 60, no segment", func(t *testing.T) {
		now := time.Now()
		experiment := Experiment{
			ID:                1,
			Name:              "test experiment",
			Uuid:              "b7fb8970-3fe7-49a7-9b3a-f670a7ae7641",
			StartDate:         now.Add(-1 * time.Hour).Unix(),
			EndDate:           now.Add(1 * time.Hour).Unix(),
			HashAttributeID:   1,
			PopulationSize:    60,
			Strategy:          "percentage_split",
			HashAttributeName: "user_id",
			Status:            ExperimentStatusRunning,
			Variants: []ExperimentVariant{
				{
					ID:                1,
					Name:              "test variant",
					TrafficAllocation: 50,
					Parameters: []ExperimentVariantParameter{
						{
							ID:                1,
							ParameterDataType: ParameterDataTypeBoolean,
							ParameterID:       1,
							ParameterName:     "enableAuth",
							RolloutValue:      "false",
						},
					},
				},
				{
					ID:                2,
					Name:              "test variant 2",
					TrafficAllocation: 50,
					Parameters: []ExperimentVariantParameter{
						{
							ID:                2,
							ParameterDataType: ParameterDataTypeBoolean,
							ParameterID:       2,
							ParameterName:     "enableAuth",
							RolloutValue:      "true",
						},
					},
				},
			},
		}

		engine := newEngine(slog.New(slog.NewTextHandler(os.Stdout, nil)))
		inPopulation := make([]int, 0)
		notPopulation := make([]int, 0)
		total := 200
		for i := 1; i <= total; i++ {
			userID := i
			attribute := NewAttribute().SetNumber("user_id", float64(userID))
			_, _, ok := engine.evaluateExperiment(&experiment, attribute, "enableAuth")
			if ok {
				inPopulation = append(inPopulation, i)
			} else {
				notPopulation = append(notPopulation, i)
			}
		}
		log.Println(len(inPopulation), inPopulation)
		log.Println(len(notPopulation), notPopulation)
		require.Equal(t, total, len(inPopulation)+len(notPopulation))
	})

}
