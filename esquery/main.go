package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"context"
	"log"
	"gopkg.in/olivere/elastic.v5"
	"text/template"
	"github.com/tehmoon/errors"
	"github.com/tehmoon/estools/esfilters/lib/esfilters"
)

func main() {
	flags := parseFlags()

	tmpl, err := template.New("root").Funcs(functionTemplates).Parse(flags.Template)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error parsing default template").Error())
	}

	client, err := elastic.NewClient(elastic.SetURL(flags.Server), elastic.SetSniff(false))
	if err != nil {
		log.Fatal(errors.Wrapf(err, "Err creating connection to server %s", flags.Server).Error())
	}

	if flags.ConfigFile != "" {
		config, err := esfilters.ImportConfigFromFile(flags.ConfigFile)
		if err != nil {
			log.Fatal(err.Error())
		}

		if flags.FilterName != "" {
			flags.QueryStringQuery, err = config.Filters.Resolve(fmt.Sprintf(`%%{filter:%s}`, flags.FilterName))
			if err != nil {
				log.Fatal(errors.Wrapf(err, "Err resolving -filter-name option").Error())
			}
		} else {
			flags.QueryStringQuery, err = config.Filters.Resolve(flags.QueryStringQuery)
			if err != nil {
				log.Fatal(errors.Wrapf(err, "Err resolving -query option").Error())
			}
		}
	}

	qs := elastic.NewQueryStringQuery(flags.QueryStringQuery)

	rq := elastic.NewRangeQuery("@timestamp").Gte(flags.From).Lte(flags.To)
	bq := elastic.NewBoolQuery().Must(qs, rq)

	res, err := client.Scroll(flags.Index).
		Query(bq).
		Sort("@timestamp", true).
		Scroll("15s").
		Size(0).
		Do(context.Background())
	if err != nil {
		if err != io.EOF {
			log.Fatalf(errors.Wrap(err, "Err querying elasticsearch").Error())
		}
	}

	scrollId := res.ScrollId
	for _, hit := range res.Hits.Hits {
		jresp := make(map[string]interface{})

		err := json.Unmarshal(*hit.Source, &jresp)
		if err != nil {
			continue
		}

		err = tmpl.Execute(os.Stdout, jresp)
		if err != nil {
			log.Fatalf(errors.Wrap(err, "Error executing template").Error())
		}
	}

	for {
		res, err := client.Scroll(flags.Index).
			Query(bq).
			Scroll("15s").
			Sort("@timestamp", true).
			ScrollId(scrollId).
			Do(context.Background())
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf(errors.Wrap(err, "Err querying elasticsearch").Error())
		}

		for _, hit := range res.Hits.Hits {
			jresp := make(map[string]interface{})
			json.Unmarshal(*hit.Source, &jresp)

			err = tmpl.Execute(os.Stdout, jresp)
			if err != nil {
				log.Fatalf(errors.Wrap(err, "Error executing template").Error())
			}
		}

		scrollId = res.ScrollId
	}

	_, err = client.ClearScroll(scrollId).
		Do(context.Background())
	if err != nil {
		log.Fatalf(errors.Wrapf(err, "Failed to clear the scrollid %s", scrollId).Error())
	}
}
