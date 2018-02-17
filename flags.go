package main

import (
	"flag"
	"os"
	"fmt"
)

type Flags struct {
	QueryStringQuery string
	StartDate string
	Server string
	Index string
	Template string
}

func parseFlags() (*Flags) {
	flags := &Flags{}

	flag.StringVar(&flags.QueryStringQuery, "query", "*", "Elasticsearch query string query")
	flag.StringVar(&flags.Server, "server", "http://localhost:9200", "Specify elasticsearch server to query")
	flag.StringVar(&flags.Index, "index", "", "Specify the elasticsearch index to query")
	flag.StringVar(&flags.Template, "template", "{{ . | json }}", "Specify Go text/template. You can use the function 'json' or 'json_indent'.")

	flag.Parse()

	if flags.Index == "" {
		fmt.Fprintln(os.Stderr, "-index is required")
		flag.Usage()
		os.Exit(2)
	}

	if flags.Template == "" {
		fmt.Fprintln(os.Stderr, "-template cannot be empty")
		flag.Usage()
		os.Exit(2)
	}

	flags.Template = fmt.Sprintf("%s\n", flags.Template)

	return flags
}

flag.Usage = func () {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	PrintDefaults()
}
