package alert

import (
	"fmt"
	"bytes"
	"os/exec"
	"github.com/tehmoon/errors"
	"encoding/json"
	"time"
	"../rules"
	"../storage"
	"github.com/olivere/elastic"
	"github.com/google/uuid"
)

type AlertConfig struct {
	Rule *rules.Rule
	Query elastic.Query
	From time.Time
	To time.Time
	Count interface{}
	Value interface{}
	ScheduledAt time.Time
}

type Alert struct {
	config *AlertConfig
	Id string
	TriggeredAt time.Time
}

type AlertPayload struct {
	From time.Time `json:"from"`
	To time.Time `json:"to"`
	ScheduledAt time.Time `json:"scheduled_at"`
	TriggeredAt time.Time `json:"triggered_at"`
	ExecutedAt time.Time `json:"executed_at"`
	Count interface{} `json:"count"`
	Value interface{} `json:"value"`
	Id string `json:"id"`
	Owners []string `json:"owners"`
	Body string `json:"body"`
	RuleName string `json:"rule_name"`
	LogURL string `json:"log_url"`
	Metadata map[string]string `json:"metadata"`
	Alert bool `json:"alert"`
}

func (a Alert) Trigger(alert bool, publicURL string) (err error) {
	ap := &AlertPayload{
		ScheduledAt: a.config.ScheduledAt,
		TriggeredAt: a.TriggeredAt,
		From: a.config.From,
		To: a.config.To,
		Count: a.config.Count,
		Value: a.config.Value,
		RuleName: a.config.Rule.Name(),
		Metadata: a.config.Rule.Metadata(),
		LogURL: a.logURL(publicURL),
		Id: a.Id,
		Owners: a.config.Rule.Owners,
		ExecutedAt: time.Now(),
		Alert: alert,
	}

	ap.Body, err = a.config.Rule.TemplateBody(ap)
	if err != nil {
		return errors.Wrap(err, "Error generating body's template")
	}

	payload, err := json.Marshal(ap)
	if err != nil {
		return errors.Wrap(err, "Error marshaling alert payload")
	}

	reader := bytes.NewReader(payload)

	cmd := exec.Command(a.config.Rule.Exec)
	cmd.Stdin = reader
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "Error executing alert command output: %s", string(output[:]))
	}

	return nil
}

func (a Alert) logURL(publicURL string) (u string) {
	if a.config.Rule.Type != rules.RuleTypeCount {
		return ""
	}

	return fmt.Sprintf("%s/alert/%s/log", publicURL, a.Id)
}

func (a Alert) Save(store StoreFunc) (err error) {
	config := &storage.WorkConfig{
		Query: a.config.Query,
		Rule: a.config.Rule,
		Id: a.Id,
	}

	return store(config)
}

type StoreFunc func(config *storage.WorkConfig) (err error)

func NewAlert(config *AlertConfig) (alert *Alert) {
	return &Alert{
		config: config,
		Id: uuid.New().String(),
		TriggeredAt: time.Now(),
	}
}
