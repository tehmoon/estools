package rules

type RuleConfig struct {
	Check string `json:"check"`
	Query string `json:"query"`
	Body string `json:"Body"`
	Name string `json:"name"`
	Log string `json:"log"`
	Metadata map[string]string `json:"metadata"`
	From *DateTimeConfig `json:"from"`
	To *DateTimeConfig `json:"to"`
	TimestampField string `json:"timestamp_field"`
	Aggregation *RuleAggregationConfig `json:"aggregation"`
	Index string `json:"index"`
	Owners []string `json:"owners"`
	Exec string `json:"exec"`
	RunEvery string `json:"run_every"`
	AlertEvery string `json:"alert_every"`
	MaxWaitSchedule string `json:"max_wait_schedule"`
	Response *RuleResponseConfig `json:"response"`
}
