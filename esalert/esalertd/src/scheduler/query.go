package scheduler

import (
	"../alert"
	"../rules"
	"../client"
	"../util"
	"context"
	"time"
	"github.com/olivere/elastic"
	"github.com/tehmoon/errors"
)

type QueryConfig struct {
	Rule *rules.Rule
	From time.Time
	To time.Time
	ClientManager *client.Manager
	Query elastic.Query
	AlertManager *alert.Manager
	ScheduledAt time.Time
}

func queryAggregationTerms(config *QueryConfig) (err error) {
	// TODO: work on this
	size := 1000
	partition := 1

	agg := elastic.NewTermsAggregation().
		Size(size).
		Partition(partition).
		Field(config.Rule.Aggregation.Field)

	search := config.ClientManager.Search(config.Rule.Index).
		Query(config.Query).
		Size(0).
		Aggregation("root", agg)

	res, err := search.Do(context.Background())
	if err != nil {
		return errors.Wrap(err, "Error querying aggregation terms")
	}

	resAgg, found := res.Aggregations.Terms("root")
	if !found {
		return errors.Wrapf(err, "Could not find aggregation named %q", "root")
	}

	for _, bucket := range resAgg.Buckets {
		value, ok := bucket.Key.(string)
		if !ok {
			continue
		}

		err = annalyze(config, bucket.DocCount, value)
		if err != nil {
			return errors.Wrap(err, "Error annalyzing query results")
		}
	}

	return nil
}

func annalyze(qc *QueryConfig, count, value interface{}) (err error) {
	check, err := qc.Rule.TemplateCheck(count)
	if err != nil {
		return errors.Wrap(err, "Error running check template")
	}

	if check != "true" {
		ac := &alert.AlertConfig{
			Rule: qc.Rule,
			Query: qc.Query,
			ScheduledAt: qc.ScheduledAt,
			From: qc.From,
			To: qc.To,
			Count: count,
			Value: value,
		}

		util.Printf("Triggering rule: %s id %s\n", qc.Rule.Name(), qc.Rule.Id())
		qc.AlertManager.Trigger(ac)
	}

	return nil
}


func queryCount(config *QueryConfig) (err error) {
	search := config.ClientManager.Search(config.Rule.Index).
		Query(config.Query).
		Size(0)

	res, err := search.Do(context.Background())
	if err != nil {
		return errors.Wrap(err, "Error querying aggregation terms")
	}

	err = annalyze(config, res.Hits.TotalHits, res.Hits.TotalHits)
	if err != nil {
		return errors.Wrap(err, "Error annalyzing query results")
	}

	return nil
}

func query(config *QueryConfig) (err error) {
	switch config.Rule.Type {
		case rules.RuleTypeCount:
			return queryCount(config)
		case rules.RuleTypeAggregationTerms:
			return queryAggregationTerms(config)
	}

	util.Fatal("Path should not be reached, please open an issue.")
	return nil
}
