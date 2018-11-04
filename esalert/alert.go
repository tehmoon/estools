package main

import (
	"io"
	"compress/gzip"
	"context"
	"log"
	"gopkg.in/olivere/elastic.v5"
	"github.com/google/uuid"
	"os/exec"
	"encoding/json"
	"time"
	"github.com/tehmoon/errors"
	"bytes"
)

type Alert struct {
	Body string `json:"body"`
	Metadata AlertMetadata `json:"metadata"`
	Id string `json:"id"`
	Log []byte `json:"-"`
	rule *Rule
}

type AlertMetadata struct {
	Query string `json:"query"`
	File string `json:"file"`
	TotalHits int64 `json:"totalHits"`
	Name string `json:"name"`
	Owners []string `json:"owners"`
	From time.Time `json:"from"`
	To time.Time `json:"to"`
	Index string `json:"index"`
	TimestampField string `json:"timestamp_field"`
}

func triggerAlert(exe string, alert *Alert) (error) {
	payload, err := json.Marshal(alert)
	if err != nil {
		return errors.Wrap(err, "Error marshaling alert to json")
	}

	reader := bytes.NewReader(payload)

	cmd := exec.Command(exe)
	cmd.Stdin = reader
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "Error executing alert command; output: %s", string(cmdOutput[:]))
	}

	return nil
}

func newAlert(rule *Rule, totalHits int64, gte, lt *time.Time) (*Alert) {
	alert := &Alert{
		Metadata: AlertMetadata{
			File: rule.file,
			Query: rule.Query,
			TotalHits: totalHits,
			Name: rule.Name,
			Owners: rule.owners,
			From: *gte,
			To: *lt,
			Index: rule.Index,
			TimestampField: rule.TimestampField,
		},
		Id: uuid.New().String(),
		Log: make([]byte, 0),
		rule: rule,
	}

	return alert
}

func (a *Alert) SaveLogs(client *elastic.Client) (error) {
	qs := elastic.NewQueryStringQuery(a.Metadata.Query)

	rq := elastic.NewRangeQuery(a.Metadata.TimestampField).
					Gte(a.Metadata.From.UTC().Format(time.RFC3339)).
					Lt(a.Metadata.To.UTC().Format(time.RFC3339))
	bq := elastic.NewBoolQuery().Must(qs, rq)

	res, err := client.Scroll(a.Metadata.Index).
		Query(bq).
		Sort("@timestamp", true).
		Scroll("5s").
		Size(500).
		Do(context.Background())
	if err != nil {
		if err != io.EOF {
			return errors.Wrap(err, "Err querying elasticsearch")
		}
	}

	if res == nil {
		return nil
	}

	lines := make([]string, 0)
	line := ""

	scrollId := res.ScrollId
	for _, hit := range res.Hits.Hits {
		line, err = processHitTemplate(hit.Source, a.rule)
		if err != nil {
			continue
		}

		lines = append(lines, line)
	}

	for {
		res, err := client.Scroll(a.Metadata.Index).
			ScrollId(scrollId).
			Do(context.Background())
		if err != nil {
			if err == io.EOF {
				break
			}

			return errors.Wrap(err, "Err querying elasticsearch when scrolling")
		}

		for _, hit := range res.Hits.Hits {
			line, err = processHitTemplate(hit.Source, a.rule)
			if err != nil {
				continue
			}

			lines = append(lines, line)
		}

		scrollId = res.ScrollId
	}

	clearScroll(client, scrollId)

	buff := bytes.NewBuffer(nil)
	writer := gzip.NewWriter(buff)

	for _, line := range lines {
		writer.Write([]byte(line))
	}

	writer.Close()
	a.Log = append(a.Log, buff.Bytes()...)

	return nil
}

func clearScroll(client *elastic.Client, id string) {
	if id != "" {
		_, err := client.ClearScroll(id).Do(context.Background())
		if err != nil {
			log.Printf("%s\n", errors.Wrap(err, "Failed to clear scroll"))
		}
	}
}

func processHitTemplate(source *json.RawMessage, rule *Rule) (string, error) {
	v := make(map[string]interface{})

	err := json.Unmarshal(*source, &v)
	if err != nil {
		return "", errors.Wrap(err, "Error marshaling JSON to map[string]internface{}")
	}

	line, err := rule.ExecTemplate(RuleTemplateLog, v)
	if err != nil {
		return "", errors.Wrap(err, "Error executing template")
	}

	return line, nil
}
