package rules

import (
	"time"
	"text/template"
	"github.com/tehmoon/errors"
	"../util"
)

type RuleResponseConfig struct {
	// Defaults to 60s
	// 0 or negative means it does not expire
	Expire string `json:"expire"`
	Tags []string `json:"tags"`
	RawArgs []string `json:"args"`
	Action string `json:"action"`
}

type RuleResponse struct {
	Expire time.Duration
	Tags []string
	Args []*template.Template
	Action string
}

func NewRuleResponse(config *RuleResponseConfig) (rr *RuleResponse, err error) {
	rr = &RuleResponse{
		Action: config.Action,
		Tags: config.Tags,
		Args: make([]*template.Template, 0),
	}

	if config.Action == "" {
		return nil, errors.Errorf("%q field cannot be empty", "action")
	}

	if len(config.Tags) == 0 {
		return nil, errors.Errorf("%q field doesn't have any tag", "tags")
	}

	expire := config.Expire

	if expire == "" {
		expire = "60s"
	}

	rr.Expire, err = time.ParseDuration(config.Expire)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing the %q field", "expire")
	}

	for _, arg := range config.RawArgs {
		tmpl, err := util.NewTemplate().Parse(arg)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing template in arg %q", arg)
		}

		rr.Args = append(rr.Args, tmpl)
	}

	return rr, nil
}
