package main

import (
	"flag"
	"os"
	"fmt"
)

type Flags struct {
	QueryStringQuery string
	Server string
	Index string
	Template string
	ConfigFile string
	FilterName string
	Tail bool
	Start string
	End string
}

func parseFlags() (*Flags) {
	flags := &Flags{}

	flag.BoolVar(&flags.Tail, "tail", false, "Keep scrolling on new data. Cannot be used with \"-end\" flag")
	flag.StringVar(&flags.Start, "start", "", "Specify when to start fetching. Elasticserach date format. Defaults to \"now\" when \"-tail\" is set")
	flag.StringVar(&flags.End, "end", "", "Specify when to end fetching. Elasticserach date format. Cannot be used with \"-tail\" flag")
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

	if flags.Start == "" && flags.Tail {
		flags.Start = "now"
	}

	if flags.Tail && flags.End != "" {
		fmt.Fprintln(os.Stderr, "-end and -tail are mutually exclusive")
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
