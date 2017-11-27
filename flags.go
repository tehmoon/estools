package main

import (
  "flag"
  "os"
  "fmt"
)

func parseFlags() (*Flags) {
  flags := &Flags{}

  flag.StringVar(&flags.QueryStringQuery, "query", "*", "Elasticsearch query string query")
  flag.StringVar(&flags.Server, "server", "http://localhost:9200", "Specify elasticsearch server to query")
  flag.StringVar(&flags.Index, "index", "", "Specify the elasticsearch index to query")

  flag.Parse()

  if flags.Index == "" {
    fmt.Fprintln(os.Stderr, "-index is required")
    flag.Usage()
    os.Exit(2)
  }

  return flags
}
