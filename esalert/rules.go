package main

import (
	"github.com/tehmoon/errors"
	"text/template"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"path/filepath"
)

var (
	functionTemplatesText = template.FuncMap{
		"newline": func() (string) {
			return "\n"
		},
		"json": func(d interface{}) (string) {
			payload, err := json.Marshal(d)
			if err != nil {
				return ""
			}

			return string(payload[:])
		},
		"json_indent": func(d interface{}) (string) {
			payload, err := json.MarshalIndent(d, "", "  ")
			if err != nil {
				return ""
			}

			return string(payload[:])
		},
	}
)

type Rules []Rule

type RuleTemplateType int
const (
	RuleTemplateCheck RuleTemplateType = iota
	RuleTemplateBody
	RuleTemplateLog
)

// Not thread safe
type Rule struct {
	Check string `json:"check"`
	Query string `json:"query"`
	Body string `json:"Body"`
	Name string `json:"name"`
	Log string `json:"log"`
	From RuleDatetime `json:"from"`
	To RuleDatetime `json:"to"`
	TimestampField string `json:"timestamp_field"`
	tCheck *template.Template
	tBody *template.Template
	tLog *template.Template
	file string
	buff *bytes.Buffer
	owners []string
	Index string `json:"-"`
}

func (r *Rule) Validate() (error) {
	if r.Check == "" {
		return errors.Errorf("Field %q is missing", "check")
	}

	if r.Log == "" {
		r.Log = `{{ . | json }}{{ newline }}`
	}

	if r.Body == "" {
		return errors.Errorf("Field %q is missing", "Body")
	}

	if r.Query == "" {
		return errors.Errorf("Field %q is missing", "query")
	}

	if r.TimestampField == "" {
		r.TimestampField = "@timestamp"
	}

	var err error
	r.tCheck, err = template.New("root").Funcs(functionTemplatesText).Parse(r.Check)
	if err != nil {
		return errors.Wrapf(err, "Bad template for field %q", "check")
	}

	r.tBody, err = template.New("root").Funcs(functionTemplatesText).Parse(r.Body)
	if err != nil {
		return errors.Wrapf(err, "Bad template for field %q", "body")
	}

	r.tLog, err = template.New("root").Funcs(functionTemplatesText).Parse(r.Log)
	if err != nil {
		return errors.Wrapf(err, "Bad template for field %q", "log")
	}

	// 60 sec period + 15 sec index time
	if r.From.Date == "" {
		r.From.Date = "now"
		r.From.Minus = "90s"
		r.From.Round = "minute"
	}

	// 15 sec index time
	if r.To.Date == "" {
		r.To.Date = "now"
		r.To.Minus = "15s"
		r.To.Round = "minute"
	}

	return nil
}

func (r Rule) ExecTemplate(t RuleTemplateType, v interface{}) (string, error) {
	r.buff.Reset()

	var err error
	switch t {
		case RuleTemplateCheck:
			err = r.tCheck.Execute(r.buff, v)
			if err != nil {
				err = errors.Wrap(err, "Error executing check template")
			}

		case RuleTemplateBody:
			err = r.tBody.Execute(r.buff, v)
			if err != nil {
				err = errors.Wrap(err, "Error executing body template")
			}

		case RuleTemplateLog:
			err = r.tLog.Execute(r.buff, v)
			if err != nil {
				err = errors.Wrap(err, "Error executing log template")
			}
	}

	return r.buff.String(), err
}

func loadRule(f string, owners []string) (*Rule, error) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to read file")
	}

	rule := &Rule{
		file: f,
		buff: bytes.NewBuffer(nil),
		owners: owners,
	}

	err = json.Unmarshal(data, &rule)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to unmarshal rule for file %q", f)
	}

	err = rule.Validate()
	if err != nil {
		return nil, errors.Wrapf(err, "Error validating rule for file %q", f)
	}

	return rule, nil
}

func loadRules(flags *Flags) (Rules, error) {
	files, err := filepath.Glob(filepath.Join(flags.Dir, "*.json"))
	if err != nil {
		return nil, errors.Wrapf(err, "Err calling %q", "filepath.Glob")
	}

	rules := make(Rules, 0)

	for _, file := range files {

		rule, err := loadRule(file, flags.Owners)
		if err != nil {
			return nil, errors.Wrap(err, "Error loading rule")
		}

		rule.Index = flags.Index
		rules = append(rules, *rule)
	}

	return rules, nil
}
