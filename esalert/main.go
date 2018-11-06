package main

import (
	"time"
	"io"
	"context"
	"log"
	"gopkg.in/olivere/elastic.v5"
	"github.com/tehmoon/errors"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
}

func main() {
	flags := parseFlags()

	client, err := elastic.NewClient(elastic.SetURL(flags.Server), elastic.SetSniff(false))
	if err != nil {
		log.Fatal(errors.Wrapf(err, "Err creating connection to server %s", flags.Server).Error())
	}

	rules, err := loadRules(flags)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error load rules").Error())
	}

	alertIndex := make(map[string]*Alert, 0)
	err = startHTTP(flags.Listen, alertIndex)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to start HTTP server").Error())
	}

	for {
		for _, rule := range rules {
			log.Printf("Running query from file %q\n", rule.file)

			gte, lt, totalHits, err := fetch(client, &rule, flags)
			if err != nil {
				log.Println(errors.Wrapf(err, "Querying elasticserach for file %q").Error())
				continue
			}

			output, err := rule.ExecTemplate(RuleTemplateCheck, totalHits)
			if err != nil {
				log.Println(errors.Wrapf(err, "Error in file %q", rule.file).Error())
				continue
			}

			if output != "true" {
				alert := newAlert(&rule, totalHits, gte, lt)

				alert.Body, err = rule.ExecTemplate(RuleTemplateBody, &alert.Metadata)
				if err != nil {
					log.Println(errors.Wrapf(err, "Error in file %q", rule.file).Error())
					continue
				}

				err = triggerAlert(flags.Exec, alert)
				if err != nil {
					log.Println(errors.Wrapf(err, "Error triggering alert in file %q", rule.file).Error())
					continue
				}

				err = alert.SaveLogs(client)
				if err != nil {
					log.Println(errors.Wrapf(err, "Error saving logs for file %q", rule.file).Error())
					continue
				}

				alertIndex[alert.Id] = alert

				log.Println(errors.Wrapf(err, "Alert for file %q output %q", rule.file, output).Error())
			}

			log.Printf("Done executing file %q\n", rule.file)
		}

		time.Sleep(time.Duration(flags.SleepFor) * time.Second)
	}
}

func fetch(client *elastic.Client, rule *Rule, flags *Flags) (*time.Time, *time.Time, int64, error) {
	gte, _ := rule.From.Time()
	lt, _ := rule.To.Time()

	qs := elastic.NewQueryStringQuery(rule.Query)
	rq := elastic.NewRangeQuery(rule.TimestampField).
					Gte(gte.UTC().Format(time.RFC3339)).
					Lt(lt.UTC().Format(time.RFC3339))
	bq := elastic.NewBoolQuery().Must(qs, rq)

	res, err := client.Search(rule.Index).
		Query(bq).
		Size(0).
		Do(context.Background())
	if err != nil {
		if err != io.EOF {
			return nil, nil, 0, errors.Wrapf(err, "Error querying from file %q", rule.file)
		}
	}

	var totalHits int64 = 0

	if res != nil {
		totalHits = res.Hits.TotalHits
	}

	return gte, lt, totalHits, nil
}
