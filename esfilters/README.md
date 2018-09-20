# esfilters

Tired of having to type over and over complex query strings in elasticsearch?
What about using dynamic 100% customisable, sharable, nestable, queries?

For example using the filters config:

```
{
  "filters": {
    "filebeat": "beat.name: \"filebeat\"",
    "production": "fields.env: \"production\"",
    "syslog": "fields.type: \"syslog\" AND %{filter:filebeat}"
  }
}
```

The query:

```
$> esfilters -c config.json resolve filter -query '%{filter:production} AND %{filter:syslog}'
```

would give you:

```
((fields.env: "production") AND (fields.type: "syslog" AND (beat.name: "filebeat")))
```

It also works on aggregations too.
Let's see an example here:

```
  - count:
    - aggregation: '{"sum": { "field": "%{placeholder:field}" }}'
      placeholders:
        - field: "@timestamp"
  - field_uniq_per_terms:
    - aggregation: '{"terms":{"field":"%{placeholder:field1}"},"aggregations":{"1":{"cardinality":{"field":"%{placeholder:field2}"}}}}'
      placeholders:
        - field1: ""
        - field2: ""
```

If you want to count all the documents that have a `@timestamp` field you can call directly `count`.
If you want to count the number of different terms you can call `field_uniq_per_terms`. In that case `field1` and `field2` are required.

`esfilters` comes with already registered aggregation filters, but you can override all of them.


## How to install
There are two ways you could install `esfilters`:

  - By going to the [release page](https://github.com/tehmoon/esfilters/releases)
  - By compiling it yourself:
    - Download and install [go](https://golang.org)
    - `git clone https://github.com/tehmoon/esfilters`
    - Set `GOPATH`: `mkdir ~/work/go && export GOPATH=~/work/go`
    - `cd esfilters`
    - Get dependencies `go get ./...`
    - Generate binary `go build`
    - Optional: Move the binary to `${GOPATH}/bin`: `go install`

## Usage

```
Usage: ./esfilters <Options> <Command> <Module> [-h | --help]
Options:
  -c string
    	Config file to use

Command:
  resolve
  add
  delete
  list

Module:
  filter
```

## Features

  - [x] Query Filters: Use filters to build a query string
  - [ ] Aggregation Filters: Use filters to build an aggregation
  - [ ] JSON Filters: Use filters to parse the elastic response
  - [x] Config file storage: Use config file to store all the filters locally
  - [x] CLI config tool: Manage the config file using the CLI
  - [x] Go library: Use this libray in all your project that could query elasticsearch
