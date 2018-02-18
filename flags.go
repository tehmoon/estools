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
	ConfigFile string
	FilterName string
}

func parseFlags() (*Flags) {
	flags := &Flags{}

	flag.StringVar(&flags.QueryStringQuery, "query", "*", "Elasticsearch query string query")
	flag.StringVar(&flags.FilterName, "filter-name", "", "If specified use the esfilter's filter as the query")
	flag.StringVar(&flags.ConfigFile, "config", "", "Use configuration file created by esfilters")
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

	if flags.FilterName != "" && flags.ConfigFile == "" {
		fmt.Fprintln(os.Stderr, "when -filter-name is used, -config has to be specified")
		flag.Usage()
		os.Exit(2)
	}

	if flags.FilterName != "" && (flags.QueryStringQuery != "*" && flags.QueryStringQuery != "") {
		fmt.Fprintln(os.Stderr, "-filter-name and -query are mutually exclusive")
		flag.Usage()
		os.Exit(2)
	}

	flags.Template = fmt.Sprintf("%s\n", flags.Template)

	return flags
}

func init() {
	flag.Usage = func () {
		fmt.Fprintf(os.Stderr, "Usage of %s: [-config=file] [-query=Query | <-config=file> <-filter-name=FilterName>] <-server=Url> <-index=Index> [-template=Template]\n", os.Args[0])
		flag.PrintDefaults()
	}
}
