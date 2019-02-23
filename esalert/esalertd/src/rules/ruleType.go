package rules

type RuleType int

const (
	RuleTypeCount = iota

	// Aggregations type are below
	RuleTypeAggregationTerms
)
