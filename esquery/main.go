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

	rq := elastic.NewRangeQuery(flags.TimestampField).Gte(flags.From).Lt(flags.To)
	bq := elastic.NewBoolQuery().Must(qs, rq)

	res, err := client.Scroll(flags.Index).
		Query(bq).
		Sort(flags.Sort, flags.Asc).
		Scroll("15s").
		Size(flags.ScrollSize).
		Do(context.Background())
	if err != nil {
		if err != io.EOF {
			log.Fatalf(errors.Wrap(err, "Err querying elasticsearch").Error())
		}
	}

	if flags.CountOnly {
		var totalHits int64 = 0

		if res != nil {
			totalHits = res.Hits.TotalHits
		}

		err = tmpl.Execute(os.Stdout, totalHits)
		if err != nil {
			log.Fatalf(errors.Wrap(err, "Error executing template").Error())
		}

		return
	}

	if res == nil {
		return
	}

	counter := 0

	scrollId := res.ScrollId
	for _, hit := range res.Hits.Hits {
		if counter == flags.Size && counter != 0 {
			break
		}

		jresp := make(map[string]interface{})

		err := json.Unmarshal(*hit.Source, &jresp)
		if err != nil {
			continue
		}

		err = tmpl.Execute(os.Stdout, jresp)
		if err != nil {
			log.Fatalf(errors.Wrap(err, "Error executing template").Error())
		}

		counter++
	}

	LOOP: for {
		res, err := client.Scroll(flags.Index).
			Query(bq).
			Scroll("15s").
			ScrollId(scrollId).
			Do(context.Background())
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf(errors.Wrap(err, "Err querying elasticsearch").Error())
		}

		for _, hit := range res.Hits.Hits {
			if counter == flags.Size && counter != 0 {
				scrollId = res.ScrollId
				break LOOP
			}

			jresp := make(map[string]interface{})
			json.Unmarshal(*hit.Source, &jresp)

			err = tmpl.Execute(os.Stdout, jresp)
			if err != nil {
				log.Fatalf(errors.Wrap(err, "Error executing template").Error())
			}

			counter++
		}

		scrollId = res.ScrollId
	}

	_, err = client.ClearScroll(scrollId).
		Do(context.Background())
	if err != nil {
		log.Fatalf(errors.Wrapf(err, "Failed to clear the scrollid %s", scrollId).Error())
	}
}
