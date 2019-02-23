package rules

import (
	"github.com/olivere/elastic"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"github.com/tehmoon/errors"
	"text/template"
	"../util"
	"sync"
	"time"
	"crypto/sha256"
	"encoding/hex"
)

type Rule struct {
	path string
	config *RuleConfig
	tCheck *template.Template
	tBody *template.Template
	tLog *template.Template
	tQuery *template.Template
	to *DateTime
	from *DateTime
	Type RuleType
	Index string
	Owners []string
	Exec string
	lastScheduled *time.Time
	RunEvery time.Duration
	MaxWaitSchedule time.Duration
	Aggregation *RuleAggregation
	Response *RuleResponse
	AlertEvery time.Duration
	id string
	sync.Mutex
}

func loadRule(f, index, exec string, owners []string) (rule *Rule, err error) {
	data, err := ioutil.ReadFile(filepath.Clean(f))
	if err != nil {
		return nil, errors.Wrap(err, "Fail to read file")
	}

	sum := sha256.Sum256(data)

	config := &RuleConfig{}

	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, errors.Wrapf(err, "Error unmarshal JSON in file %q", f)
	}

	if config.Check == "" {
		return nil, errors.Errorf("Field %q is missing", "check")
	}

	if config.Body == "" {
		return nil, errors.Errorf("Field %q is missing", "Body")
	}

	if config.Owners == nil {
		config.Owners = owners
	}

	if config.Exec == "" {
		config.Exec = exec
	}

	if config.Index == "" {
		config.Index = index
	}

	if config.Index == "" {
		config.Index = index
	}

	if config.TimestampField == "" {
		config.TimestampField = "@timestamp"
	}

	if config.Metadata == nil {
		config.Metadata = make(map[string]string)
	}

	if config.Log == "" {
		config.Log = `{{ . | json }}{{ newline }}`
	}

	if config.RunEvery == "" {
		config.RunEvery = "45s"
	}

	if config.AlertEvery == "" {
		config.AlertEvery = "0s"
	}

	// Always after config.RunEvery
	if config.MaxWaitSchedule == "" {
		config.MaxWaitSchedule = config.RunEvery
	}

	if config.To == nil {
		config.To = &DateTimeConfig{
			Date: "now",
			Round: "minute",
		}
	}

	if config.From == nil {
		config.From = &DateTimeConfig{
			Date: "now",
			Minus: "60s",
			Round: "minute",
		}
	}

	return NewRule(f, hex.EncodeToString(sum[:]), config)
}

func NewRule(f, sum string, config *RuleConfig) (rule *Rule, err error) {
	rule = &Rule{
		path: f,
		config: config,
		Type: RuleTypeCount,
		Owners: config.Owners,
		Exec: config.Exec,
		Index: config.Index,
		id: sum,
	}

	rule.RunEvery, err = time.ParseDuration(config.RunEvery)
	if err != nil {
		return nil, errors.Wrapf(err, "Bad duration for field %q", "run_every")
	}

	rule.AlertEvery, err = time.ParseDuration(config.AlertEvery)
	if err != nil {
		return nil, errors.Wrapf(err, "Bad duration for field %q", "alert_every")
	}

	rule.MaxWaitSchedule, err = time.ParseDuration(config.MaxWaitSchedule)
	if err != nil {
		return nil, errors.Wrapf(err, "Bad duration for field %q", "max_wait_schedule")
	}

	if rule.MaxWaitSchedule > rule.RunEvery {
		rule.MaxWaitSchedule = rule.RunEvery
	}

	if rule.Exec == "" {
		return nil, errors.New("No defined executable in either global configuration or rule configuration")
	}

	rule.tCheck, err = util.NewTemplate().Parse(config.Check)
	if err != nil {
		return nil, errors.Wrapf(err, "Bad template for field %q", "check")
	}

	rule.tBody, err = util.NewTemplate().Parse(config.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Bad template for field %q", "body")
	}

	rule.tLog, err = util.NewTemplate().Parse(config.Log)
	if err != nil {
		return nil, errors.Wrapf(err, "Bad template for field %q", "log")
	}

	rule.tQuery, err = util.NewTemplate().Parse(config.Query)
	if err != nil {
		return nil, errors.Wrapf(err, "Bad template for field %q", "query")
	}

	rule.from, err = NewDateTime(config.From)
	if err != nil {
		return nil, errors.Wrapf(err, "Error validating %q field", "from")
	}

	rule.to, err = NewDateTime(config.To)
	if err != nil {
		return nil, errors.Wrapf(err, "Error validating %q field", "to")
	}

	now := time.Now()
	if rule.from.Time(&now).UnixNano() > rule.to.Time(&now).UnixNano() {
		return nil, errors.Errorf("%s field is greater than %s field", "from", "to")
	}

	if rule.config.Aggregation != nil {
		rule.Aggregation, err = NewRuleAggregation(rule.config.Aggregation)
		if err != nil {
			return nil, errors.Wrapf(err, "Error validating %q field", "aggregation")
		}

		rule.Type = rule.Aggregation.Type
	}

	if rule.config.Response != nil {
		rule.Response, err = NewRuleResponse(rule.config.Response)
		if err != nil {
			return nil, errors.Wrapf(err, "Error validating %q field", "response")
		}
	}

	return rule, nil
}

func loadRules(p, index, exec string, owners []string) ([]*Rule, error) {
	files, err := filepath.Glob(filepath.Join(p, "*.json"))
	if err != nil {
		return nil, errors.Wrapf(err, "Err calling %q", "filepath.Glob")
	}

	rules := make([]*Rule, 0)

	for _, file := range files {
		rule, err := loadRule(file, index, exec, owners)
		if err != nil {
			return nil, errors.Wrap(err, "Error loading rule")
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

type TemplateQueryRoot struct {
	From string
	To string
}

func (r Rule) Name() (name string) {
	return r.path
}

func (r Rule) TemplateBody(v interface{}) (body string, err error) {
	return util.TemplateToString(r.tBody, v)
}

func (r Rule) TemplateLog(v interface{}) (data []byte, err error) {
	return util.TemplateToBytes(r.tLog, v)
}

func (r Rule) TemplateQuery(from, to time.Time) (query string, err error) {
	root := &TemplateQueryRoot{
		From: util.FormatTime(from),
		To: util.FormatTime(to),
	}

	return util.TemplateToString(r.tQuery, root)
}

func (r Rule) TemplateCheck(v interface{}) (text string, err error) {
	return util.TemplateToString(r.tCheck, v)
}

func (r *Rule) GenerateQuery(from, to time.Time) (bq elastic.Query, err error) {
	query, err := r.TemplateQuery(from, to)
	if err != nil {
		return nil, errors.Wrap(err, "Error generating query from template")
	}

	qs := elastic.NewQueryStringQuery(query)
	rq := elastic.NewRangeQuery(r.config.TimestampField).
		Gte(util.FormatTime(from)).
		Lt(util.FormatTime(to))
	bq = elastic.NewBoolQuery().Must(qs, rq)

	return bq, nil
}

func (r *Rule) From(now *time.Time) (time.Time) {
	return r.from.Time(now)
}

func (r Rule) To(now *time.Time) (time.Time) {
	return r.to.Time(now)
}

func (r Rule) Metadata() (map[string]string) {
	return r.config.Metadata
}

func (r Rule) Id() (id string) {
	return r.id
}
