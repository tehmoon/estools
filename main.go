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
  "github.com/tehmoon/errors"
)

type Flags struct {
  QueryStringQuery string
  StartDate string
  Server string
  Index string
}

type JsonResponse struct {
  Timestamp string `json:"@timestamp"`
  Message string `json:"message"`
}

func main() {
  flags := parseFlags()

  client, err := elastic.NewClient(elastic.SetURL(flags.Server), elastic.SetSniff(false))
  if err != nil {
    log.Fatal(errors.Wrapf(err, "Err creating connection to server %s", flags.Server).Error())
  }

  qs := elastic.NewQueryStringQuery(flags.QueryStringQuery)

  var scrollId string
  var rq *elastic.RangeQuery
  var bq *elastic.BoolQuery
  var jresp *JsonResponse

  for {
    res, err := client.Search(flags.Index).
      Query(qs).
      Size(1).
      Sort("@timestamp", false).
      Do(context.Background())
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error querying elasticserach cluster: %v")
      os.Exit(2)
    }

    jresp = &JsonResponse{}

    if len(res.Hits.Hits) != 0 {
      json.Unmarshal(*res.Hits.Hits[0].Source, jresp)
      break
    }

    log.Println("No results found... Sleeping")
    time.Sleep(5 * time.Second)
    continue
  }

  for {
    rq = elastic.NewRangeQuery("@timestamp").Gt(jresp.Timestamp)
    bq = elastic.NewBoolQuery().Must(qs, rq)

    res, err := client.Scroll(flags.Index).
      Query(bq).
      Sort("@timestamp", false).
      Scroll("15s").
      Size(0).
      Do(context.Background())
    if err != nil {
      if err == io.EOF {
        time.Sleep(5 * time.Second)
        continue
      }

      log.Fatalf("Err querying elasticsearch. Error: %v", err)
    }

    scrollId = res.ScrollId
      for _, hit := range res.Hits.Hits {
        jresp := &JsonResponse{}
        json.Unmarshal(*hit.Source, jresp)
        fmt.Println(jresp.Message)
      }
  jresp = &JsonResponse{}

  json.Unmarshal(*res.Hits.Hits[len(res.Hits.Hits) - 1].Source, jresp)

    for {
      res, err := client.Scroll(flags.Index).
        Query(bq).
	Scroll("15s").
        Sort("@timestamp", false).
        ScrollId(scrollId).
        Do(context.Background())
      if err != nil {
        if err == io.EOF {
          break
        }

        log.Fatalf(errors.Wrap(err, "Err querying elasticsearch").Error())
      }

      for _, hit := range res.Hits.Hits {
        jresp := &JsonResponse{}
        json.Unmarshal(*hit.Source, jresp)
        fmt.Println(jresp.Message)
      }
  jresp = &JsonResponse{}

  json.Unmarshal(*res.Hits.Hits[len(res.Hits.Hits) - 1].Source, jresp)
  fmt.Println(jresp)

      scrollId = res.ScrollId
    }

    _, err = client.ClearScroll(scrollId).
      Do(context.Background())
    if err != nil {
      log.Fatalf(errors.Wrapf(err, "Failed to clear the scrollid %s", scrollId).Error())
    }
  }
}
