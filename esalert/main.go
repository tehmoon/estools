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

	for {
		for _, rule := range rules {
			log.Printf("Running query from file %q\n", rule.file)

			totalHits, err := fetchTotalHits(client, &rule, flags)
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
				alert := newAlert(&rule, totalHits)

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

				log.Println(errors.Wrapf(err, "Alert for file %q output %q", rule.file, output).Error())
			}

			log.Printf("Done executing file %q\n", rule.file)
		}

		time.Sleep(time.Duration(flags.SleepFor) * time.Second)
	}
}

func fetchTotalHits(client *elastic.Client, rule *Rule, flags *Flags) (int64, error) {
	qs := elastic.NewQueryStringQuery(rule.Query)

	// TODO: make flag
	rq := elastic.NewRangeQuery("@timestamp").Gte("now-1d/d")
	bq := elastic.NewBoolQuery().Must(qs, rq)

	res, err := client.Search(flags.Index).
		Query(bq).
		Size(0).
		Do(context.Background())
	if err != nil {
		if err != io.EOF {
			return 0, errors.Wrapf(err, "Error querying from file %q", rule.file)
		}
	}

	var totalHits int64 = 0

	if res != nil {
		totalHits = res.Hits.TotalHits
	}

	return totalHits, nil
}
