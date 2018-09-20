# Esquery

Because we all need a `tail -f` on elasticsearch logs.

It uses the Go template package from `text/template` which means you can create powerful template to customize the output
from the JSON response of Elasticsearch.

Estail also uses [esfilters](https://github.com/tehmoon/estools/esfilters) which enables you to save your queries easily.

It also sort on the field `@timestamp` only.

## How to contribute

File an issue or a PR it's more than welcomed

## Help

```
Usage of ./estail: [-config=file] [-query=Query | <-config=file> <-filter-name=FilterName>] <-server=Url> <-index=Index> [-template=Template]
  -config string
    	Use configuration file created by esfilters
  -filter-name string
    	If specified use the esfilter's filter as the query
  -index string
    	Specify the elasticsearch index to query
  -query string
    	Elasticsearch query string query (default "*")
  -server string
    	Specify elasticsearch server to query (default "http://localhost:9200")
  -sort string
    	Field to sort on (default "@timestamp")
  -template string
    	Specify Go text/template. You can use the function 'json' or 'json_indent'. (default "{{ . | json }}")
```
