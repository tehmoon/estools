package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"context"
	"log"
	"gopkg.in/olivere/elastic.v5"
	"time"
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

	lastTimestamp, err := getLastTimestamp(client, flags.Index, qs)
	if err != nil {
		log.Fatal(err.Error())
	}

	for {
		rq := elastic.NewRangeQuery("@timestamp").Gt(lastTimestamp)
		bq := elastic.NewBoolQuery().Must(qs, rq)

		res, err := client.Scroll(flags.Index).
			Query(bq).
			Sort("@timestamp", true).
			Scroll("15s").
			Size(0).
			Do(context.Background())
		if err != nil {
			if err == io.EOF {
				time.Sleep(5 * time.Second)
				continue
			}

			log.Fatalf(errors.Wrap(err, "Err querying elasticsearch").Error())
		}

		scrollId := res.ScrollId
		for _, hit := range res.Hits.Hits {
			jresp := make(map[string]interface{})

			err := json.Unmarshal(*hit.Source, &jresp)
			if err != nil {
				continue
			}

			if timestamp, found := jresp["@timestamp"]; found {
				if timestamp, ok := timestamp.(string); ok {
					lastTimestamp = timestamp
				}
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

				if timestamp, found := jresp["@timestamp"]; found {
					if timestamp, ok := timestamp.(string); ok {
						lastTimestamp = timestamp
					}
				}

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
}

func getLastTimestamp(client *elastic.Client, index string, qs *elastic.QueryStringQuery) (string, error) {
	for {
		res, err := client.Search(index).
			Query(qs).
			Size(1).
			Sort("@timestamp", false).
			Do(context.Background())
		if err != nil {
			return "", errors.Wrap(err, "Err querying elasticserach cluster")
		}

		jresp := make(map[string]interface{})

		if len(res.Hits.Hits) != 0 {
			json.Unmarshal(*res.Hits.Hits[0].Source, &jresp)
			if timestamp, found := jresp["@timestamp"]; found {
				if timestamp, ok := timestamp.(string); ok {
					return timestamp, nil
				}
			}

			break
		}

		log.Println("No results found... Sleeping")
		time.Sleep(5 * time.Second)
		continue
	}

	return "", errors.New("Unknown error")
}
