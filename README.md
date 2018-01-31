# estail

Because we all need a `tail -f` on elasticsearch logs.

It uses the Go template package from `text/template` which means you can create powerful template to customize the output
from the JSON response of Elasticsearch.

I'm planing to port a query filter engine that needs dusting for you to be able to save all the queries you saved.

It also sort on the field `@timestamp` only.

## Help

```
Usage of ./estail:
  -index string
    	Specify the elasticsearch index to query
  -query string
    	Elasticsearch query string query (default "*")
  -server string
    	Specify elasticsearch server to query (default "http://localhost:9200")
  -template string
    	Specify Go text/template. You can use the function 'json' or 'json_indent'. (default "{{ . | json }}")
```
