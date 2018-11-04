# Esquery

Similar to `estail` but without the `tail` feature.

It uses the Go template package from `text/template` which means you can create powerful template to customize the output
from the JSON response of Elasticsearch.

Estail also uses [esfilters](https://github.com/tehmoon/estools/esfilters) which enables you to save your queries easily.

## How to contribute

File an issue or a PR it's more than welcomed

## Help

```
Flag "-index" is required
Usage of ./esquery: [-config=file] [-query=Query | <-config=file> <-filter-name=FilterName>] <-server=Url> <-index=Index> [-to=date] [-from=date] [-template=Template] [-sort=Field] [-asc] [-size=Size] [-count-only] [-scroll-size=Size]
  -asc
      Sort by asc
  -config string
      Use configuration file created by esfilters
  -count-only
      Only displays the match number
  -filter-name string
      If specified use the esfilter's filter as the query
  -from string
      Elasticsearch date for gte (default "now-15m")
  -index string
      Specify the elasticsearch index to query
  -query string
      Elasticsearch query string query (default "*")
  -scroll-size int
      Document to return between each scroll (default 500)
  -server string
      Specify elasticsearch server to query (default "http://localhost:9200")
  -size int
      Overall number of results to display, does not change the scroll size
  -sort string
      Sort field (default "@timestamp")
  -template string
      Specify Go text/template. You can use the function 'json' or 'json_indent'. (default "{{ . | json }}")
  -to string
      Elasticsearch date for lte (default "now")
```
