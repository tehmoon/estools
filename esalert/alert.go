package main

import (
	"os/exec"
	"encoding/json"
	"github.com/tehmoon/errors"
	"bytes"
)

type Alert struct {
	Body string `json:"body"`
	Metadata AlertMetadata `json:"metadata"`
}

type AlertMetadata struct {
	Query string `json:"query"`
	File string `json:"file"`
	TotalHits int64 `json:"totalHits"`
	Name string `json:"name"`
	Owners []string `json:"owners"`
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

func newAlert(rule *Rule, totalHits int64) (*Alert) {
	alert := &Alert{
		Metadata: AlertMetadata{
			File: rule.file,
			Query: rule.Query,
			TotalHits: totalHits,
			Name: rule.Name,
			Owners: rule.owners,
		},
	}

	return alert
}
