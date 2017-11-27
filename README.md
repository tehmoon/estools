# estail

Because we all need a `tail -f` on elasticsearch logs.

For now it only gets the `message` field from the source elasticseach message. I also plan to add some kind of configuration to save queries easier.

It also sort on the field `@timestamp` only.

## Help

```
Usage of estail:
  -index string
    	Specify the elasticsearch index to query
  -query string
    	Elasticsearch query string query (default "*")
  -server string
    	Specify elasticsearch server to query (default "http://localhost:9200")
```
