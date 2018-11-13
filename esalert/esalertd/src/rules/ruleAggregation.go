package rules

import (
	"github.com/tehmoon/errors"
)

type RuleAggregation struct {
	Type RuleType
	Field string
}

type RuleAggregationConfig struct {
	Type string `json:"type"`
	Field string `json:"field"`
}

func NewRuleAggregation(config *RuleAggregationConfig) (ra *RuleAggregation, err error) {
	ra = &RuleAggregation{
		Field: config.Field,
	}

	switch t := config.Type; t {
		case "terms":
			ra.Type = RuleTypeAggregationTerms
		case "":
			return nil, errors.Errorf("Missing %q field", "type")
		default:
			return nil, errors.Errorf("Aggregation type %q is not supported", t)
	}

	return ra, nil
}
