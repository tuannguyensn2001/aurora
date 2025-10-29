package solver

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"api/internal/model"

	"resty.dev/v3"
)

type CheckSegmentConflictResponse struct {
	Valid bool
}

type SolverResponse struct {
	CheckResult string `json:"check_result"`
	Model       []struct {
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
	} `json:"model"`
}

type Solver interface {
	CheckSegmentsConflict(segments []model.Segment) (*CheckSegmentConflictResponse, error)
}

type solver struct {
	endpointUrl string
}

func New(endpointUrl string) Solver {
	return &solver{
		endpointUrl: endpointUrl,
	}
}

type Attribute struct {
	Name     string
	DataType string
}

type Compose struct {
	Rules      [][]string
	Attributes map[string]Attribute
}

var templateStr = `
{{range $key, $attr := .Attributes}}(declare-const {{$attr.Name}} {{$attr.DataType}})
{{end}}

(assert
    (and
        {{range $ruleIdx, $rule := .Rules}}
        (or
            {{range $conditionIdx, $condition := $rule}}
            {{$condition}}
            {{end}}
        )
        {{end}}
    )
)
`

func (s *solver) CheckSegmentsConflict(segments []model.Segment) (*CheckSegmentConflictResponse, error) {

	if len(segments) == 0 {
		return &CheckSegmentConflictResponse{
			Valid: true,
		}, nil
	}

	str, err := s.parse(segments)
	if err != nil {
		return nil, err
	}

	rt := resty.New().SetBaseURL(s.endpointUrl).SetHeader("Content-Type", "application/json")
	resp, err := rt.R().SetBody(map[string]string{
		"constraint": str,
	}).Post("/solve")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("solver returned error: %s", resp.Status())
	}

	var result SolverResponse
	if err := json.Unmarshal([]byte(resp.String()), &result); err != nil {
		return nil, err
	}

	return &CheckSegmentConflictResponse{
		Valid: result.CheckResult != "sat",
	}, nil
}

func (s *solver) parse(segments []model.Segment) (string, error) {
	rules := make([][]string, 0)
	attributeMap := make(map[string]Attribute)
	for _, segment := range segments {
		tempAttributeMap, rulesSegment := s.parseSegment(&segment)
		for name, attribute := range tempAttributeMap {
			attributeMap[name] = attribute
		}
		rules = append(rules, rulesSegment)
	}

	compose := Compose{
		Rules:      rules,
		Attributes: attributeMap,
	}

	var buf strings.Builder
	tmpl, err := template.New("solver").Parse(templateStr)
	if err != nil {
		return "", err
	}

	if err := tmpl.Execute(&buf, compose); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *solver) parseSegment(segment *model.Segment) (map[string]Attribute, []string) {
	attributeMap := make(map[string]Attribute)
	rules := make([]string, 0)

	for _, rule := range segment.Rules {
		for _, condition := range rule.Conditions {
			attributeMap[condition.Attribute.Name] = Attribute{
				Name:     condition.Attribute.Name,
				DataType: s.getZ3Type(condition.Attribute.DataType),
			}
		}
	}

	for _, rule := range segment.Rules {
		conditions := make([]string, 0)
		for _, condition := range rule.Conditions {
			conditions = append(conditions, s.conditionToZ3(condition.Attribute.Name, condition.Attribute.DataType, condition.Operator, condition.Value))
		}
		rules = append(rules, fmt.Sprintf("(and %s)", strings.Join(conditions, " ")))
	}

	return attributeMap, rules
}

// getZ3Type converts our DataType to Z3 type string
func (s *solver) getZ3Type(dataType model.DataType) string {
	switch dataType {
	case model.DataTypeString, model.DataTypeEnum:
		return "String"
	case model.DataTypeNumber:
		return "Int"
	case model.DataTypeBoolean:
		return "Bool"
	default:
		return "String"
	}
}

// conditionToZ3 converts a condition to Z3 expression string
func (s *solver) conditionToZ3(attrName string, dataType model.DataType, operator model.ConditionOperator, value string) string {
	switch operator {
	case model.ConditionOperatorEquals:
		if dataType == model.DataTypeString || dataType == model.DataTypeEnum {
			return fmt.Sprintf(`(= %s "%s")`, attrName, value)
		}
		return fmt.Sprintf("(= %s %s)", attrName, value)

	case model.ConditionOperatorNotEquals:
		if dataType == model.DataTypeString || dataType == model.DataTypeEnum {
			return fmt.Sprintf(`(not (= %s "%s"))`, attrName, value)
		}
		return fmt.Sprintf("(not (= %s %s))", attrName, value)

	case model.ConditionOperatorGreaterThan:
		return fmt.Sprintf("(> %s %s)", attrName, value)

	case model.ConditionOperatorLessThan:
		return fmt.Sprintf("(< %s %s)", attrName, value)

	case model.ConditionOperatorGreaterThanOrEqual:
		return fmt.Sprintf("(>= %s %s)", attrName, value)

	case model.ConditionOperatorLessThanOrEqual:
		return fmt.Sprintf("(<= %s %s)", attrName, value)

	case model.ConditionOperatorIn:
		// IN operator: (or (= attr "val1") (= attr "val2") ...)
		values := strings.Split(value, ",")
		var orConditions []string
		for _, val := range values {
			val = strings.TrimSpace(val)
			if dataType == model.DataTypeString || dataType == model.DataTypeEnum {
				orConditions = append(orConditions, fmt.Sprintf(`(= %s "%s")`, attrName, val))
			} else {
				orConditions = append(orConditions, fmt.Sprintf("(= %s %s)", attrName, val))
			}
		}
		if len(orConditions) == 1 {
			return orConditions[0]
		}
		return fmt.Sprintf("(or %s)", strings.Join(orConditions, " "))

	case model.ConditionOperatorNotIn:
		// NOT IN operator: (and (not (= attr "val1")) (not (= attr "val2")) ...)
		values := strings.Split(value, ",")
		var andConditions []string
		for _, val := range values {
			val = strings.TrimSpace(val)
			if dataType == model.DataTypeString || dataType == model.DataTypeEnum {
				andConditions = append(andConditions, fmt.Sprintf(`(not (= %s "%s"))`, attrName, val))
			} else {
				andConditions = append(andConditions, fmt.Sprintf("(not (= %s %s))", attrName, val))
			}
		}
		if len(andConditions) == 1 {
			return andConditions[0]
		}
		return fmt.Sprintf("(and %s)", strings.Join(andConditions, " "))

	case model.ConditionOperatorContains:
		// CONTAINS for strings: (str.contains_char attr "val")
		// Note: Z3 uses str.contains, but might need different approach for substring matching
		return fmt.Sprintf(`(str.contains %s "%s")`, attrName, value)

	case model.ConditionOperatorNotContains:
		// NOT CONTAINS: (not (str.contains attr "val"))
		return fmt.Sprintf(`(not (str.contains %s "%s"))`, attrName, value)

	default:
		return ""
	}
}
