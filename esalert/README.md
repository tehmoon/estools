# Esalert

Wanted a simple alerting tool based on the amazing [Elasticsearch](https://www.elastic.co/products/elasticsearch)?

Search no more, `esalert` *is the tool* right for *you*.

## Design

`esalert` is at an early stage right now, but the design is rather simple in order to make it supa fast.

It's a single process, not highly available right now, that, giving a directory, will look for `json` files call `rules`.

At each iteration, `esalert` will go through all the loaded `rules` and start querying `elasticsearch`.

Depending on the `totalHits` for that query and the `rule`'s configuration, it will trigger an alert or not.

At the end or the batch of queries, it'll `-sleep-for` seconds until repeating this.

When an alert gets thrown, `esalert` will scroll through the results using the *same* set of timestamp as for the query and save the result in memory.

`esalert` also provides an HTTP API so you can easily retrieve logs for your alert. 

### Rule

A `rule` is a simple `json` file.

It has to have the following fields:

```
{
  "query": string,
  "check": string,
  "body": string,
  "log": string,
  "from": {
    "minus": string,
    "plus": string,
    "round": string,
    "date": string,
    "layout": string
  },
  "to": {
    "minus": string,
    "plus": string,
    "round": string,
    "date": string,
    "layout": string
  },
  "timestamp_field": string
}
```

#### Query: required

This is the `query string query` passed to `elasticsearch`

#### Check: required

`Check` takes the form of a [go template](https://godoc.org/text/template). It has to evaluate to a `bool` `true` otherwise an alert is triggered.

The object that is passed as the root is the `integer: totalHits`.

IE:

```
  "check": "{{ eq . 0 }}"
```

Will check if `totalHits` is equal to `0`. If not, an alert gets thrown.

#### Body: required

This is the body of the alert. Once again, it's using the [go template](https://godoc.org/text/template) so you have a lot of flexibility.

The root of the object of type `AlertMetadata`:

```
type AlertMetadata struct {
  Query string
  File string
  TotalHits int64
  Name string
  Owners []string
  From time.Time
  To time.Time
  Index string
  TimestampField string
}
```

IE:

```
  "body": "Host: localhost had \"{{ .TotalHits }}\" in the past minute"
```

#### Log: optional defaults to '{{ . | json }}{{ newline }}'

When the alert gets thrown and `esalert` has to scroll through the results, it will apply the template specified in `log` to each result.

The root of the object being the source of the elasticsearch document.

#### Timestamp_field: optional defaults to '@timestamp`

If you don't want to use the field `@timestamp` as time range for your query, you can use this field

#### From: optional defaults to {"minus": "90s", "round": "minute", "date": "now"}

In order for `esalert` to be efficient, it needs to query really fast on the `totalHits` to determine if an alert must be thrown or not.

Then it needs to gets logs. Getting logs is scrolling through the results and saving that results.

In order to get the same number of results as previously got from the query, it needs to scroll over the same date range.

This is where the `dateTime` data structure comes into play.

With this you can:

  - subtract a duration with `minus` to the `date`
  - set the `date` to any format you would like thanks to the `layout` field. See https://godoc.org/time
  - add a duration with `plus` to the `date`
  - set the date to now with the keyword `now` in the field `date`
  - round the date down to the level of your choice with the field `round`: `minute`, `nanosecond`, `day`, `week` and so on

NB: From will be inclusive as opposed to To.

#### To: optional defaults to {"minus": "15s", "round": "minute", "date": "now"}

Same as `from` except it is `<` exclusive and *not* `<=` inclusive.

### Alerting

When an alert gets thrown, the executable from the `-exec` flags is executed.

It'll write to `stdin` the alert of type `Alert` as `JSON`.

IE:

```
{
  "body": string,
  "metadata": {
    "query": string,
    "file": string,
    "totalHits": int,
    "name": string,
    "owners": []string,
    "from": string,
    "to": string,
    "index": string,
    "timestamp_field": string
  },
  "id": string,
}
```

Here is an example of a simple `shell` script that uses `curl` and `jq` to send the alert to `slack`:

```
#!/bin/sh

payload=$(cat -)
export payload

body="$(printenv payload | jq -jr '"<@" + .metadata.owners[] + "> " , .body + " http://localhost:8080/alert/" + .id')"
export body

curl -X POST -H 'Content-type: application/json' --data "{\"text\":'${body}'}" https://hooks.slack.com/services/SECRET
```

In this case, `Owners` will be `slack` handles. So if we pass `-owners blih,blah`, the output on `slack` will be like this:

```
@blih @blah Suspicious change on host: "localhost" from auditbeat
```

You are more than welcomed to re-use that script. In following versions I will try to include them directly to `esalert`.

### HTTP API

`esalert` provides an HTTP API which you can query to get moe information

#### GET /alert/{id}

When the alert gets thrown, you get an `id` field.

Querying this endpoint will give you the logs previously generated from the scroll + template in `text/plain` format so you can download and analyse the output.

## Usage

```
Usage of ./esalert:
  -dir string
      Directory where the .json files are
  -exec string
      Execute a command when alerting
  -index string
      Specify the elasticsearch index to query
  -listen string
      Start HTTP server and listen in ip:port (default ":7769")
  -owners string
      List of default owners separated by "," to notify
  -server string
      Specify elasticsearch server to query (default "http://localhost:9200")
  -sleep-for int
      Sleep for in seconds after all queries have been ran (default 60)
```

## Contributing

Issues, PRs and ideas are more than welcome. File one and I'll review it ASAP.

## Rule examples:

These are some examples of what I use in production. This might be helpful to know where to begin.

### Get soon to expire certificates

```
{
  "query": "*",
  "timestamp_field": "tlsbeat.certificate.no_after",
  "from": {
    "round": "minute",
    "date": "now"
  },
  "to": {
    "round": "week",
    "date": "now",
    "plus": "168h"
  },
  "check": "{{ eq . 0 }}",
  "body": "There are some certificate due to expire next week!"
}
```

### Make sure the number of backups where OK

```
{
  "query": "beat.name: \"filebeat\" && fields.type: \"backup_kafka_done\"",
  "check": "{{ eq . 2560 }}",
  "log": "{{ .message }}{{ newline }}",
  "to": {
    "round": "hour",
    "date": "now"
  },
  "from": {
    "round": "hour",
    "date": "now",
    "minus": "62m"
  },
  "body": "Backup kafka had more or less than 2560 hit: \"{{ .TotalHits }}\" in the past hour"
}
```

### If error.log contains new entries

```
{
  "query": "fields.type: \"cs_error\" && NOT filebeat.cs_error.msg: \"Got GOAWAY\"",
  "check": "{{ eq . 0 }}",
  "log": "{{ .message }}{{ newline }}",
  "body": "Got {{ .TotalHits }} errors in the past minute"
}
```
