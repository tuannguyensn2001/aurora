package solver

import (
	"api/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSegment(t *testing.T) {
	tests := []struct {
		name             string
		segment          *model.Segment
		expectAttributes []Attribute
		expectRules      []string
	}{
		{
			name: "test 1",
			segment: &model.Segment{
				Rules: []model.SegmentRule{
					{
						Name: "rule1",
						Conditions: []model.SegmentRuleCondition{
							{
								AttributeID: 1,
								Operator:    model.ConditionOperatorEquals,
								Value:       "VN",
								Attribute: &model.Attribute{
									Name:     "country",
									DataType: model.DataTypeString,
								},
							},
							{
								AttributeID: 2,
								Operator:    model.ConditionOperatorGreaterThanOrEqual,
								Value:       "18",
								Attribute: &model.Attribute{
									Name:     "age",
									DataType: model.DataTypeNumber,
								},
							},
						},
					},
				},
			},
			expectAttributes: []Attribute{
				{
					Name:     "country",
					DataType: "String",
				},
				{
					Name:     "age",
					DataType: "Int",
				},
			},
			expectRules: []string{
				"(and (= country \"VN\") (>= age 18))",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attributes, rules := (&solver{}).parseSegment(test.segment)
			require.Equal(t, test.expectAttributes, attributes)
			require.Equal(t, test.expectRules, rules)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		segments    []model.Segment
		expectRules string
	}{
		{
			name: "test 1",
			segments: []model.Segment{
				{
					Rules: []model.SegmentRule{
						{
							Conditions: []model.SegmentRuleCondition{
								{
									AttributeID: 1,
									Operator:    model.ConditionOperatorEquals,
									Value:       "VN",
									Attribute: &model.Attribute{
										Name:     "country",
										DataType: model.DataTypeString,
									},
								},
								{
									AttributeID: 2,
									Operator:    model.ConditionOperatorGreaterThanOrEqual,
									Value:       "18",
									Attribute: &model.Attribute{
										Name:     "age",
										DataType: model.DataTypeNumber,
									},
								},
							},
						},
						{
							Conditions: []model.SegmentRuleCondition{
								{
									AttributeID: 1,
									Operator:    model.ConditionOperatorEquals,
									Value:       "VN",
									Attribute: &model.Attribute{
										Name:     "country",
										DataType: model.DataTypeString,
									},
								},
								{
									AttributeID: 2,
									Operator:    model.ConditionOperatorEquals,
									Value:       "16",
									Attribute: &model.Attribute{
										Name:     "age",
										DataType: model.DataTypeNumber,
									},
								},
							},
						},
					},
				},
				{
					Rules: []model.SegmentRule{
						{
							Conditions: []model.SegmentRuleCondition{
								{
									AttributeID: 1,
									Operator:    model.ConditionOperatorEquals,
									Value:       "VN",
									Attribute: &model.Attribute{
										Name:     "country",
										DataType: model.DataTypeString,
									},
								},
								{
									AttributeID: 2,
									Operator:    model.ConditionOperatorLessThan,
									Value:       "18",
									Attribute: &model.Attribute{
										Name:     "age",
										DataType: model.DataTypeNumber,
									},
								},
							},
						},
					},
				},
			},
			expectRules: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rules, err := (&solver{}).parse(test.segments)
			require.NoError(t, err)
			require.Equal(t, test.expectRules, rules)
		})
	}

}
