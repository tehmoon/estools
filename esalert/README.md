# Esalert

Wanted a simple alerting tool based on the amazing [Elasticsearch](https://www.elastic.co/products/elasticsearch)?

Search no more, `esalert` *is the tool* right for *you*.

## Design

`esalert` is at an early stage right now, but the design is rather simple in order to make it supa fast.

It's a single process, not highly available right now, that, giving a directory, will look for `json` files call `rules`.

At each iteration, `esalert` will go through all the loaded `rules` and start querying `elasticsearch`.

Depending on the `totalHits` for that query and the `rule`'s configuration, it will trigger an alert or not.

At the end or the batch of queries, it'll `-sleep-for` seconds until repeating this.

### Rule

A `rule` is a simple `json` file.

It has to have the following fields:

```
{
  "query": string,
  "check": string,
  "body": string
}
```

#### Query

This is the `query string query` passed to `elasticsearch`

#### Check

`Check` takes the form of a [go template](https://godoc.org/text/template). It has to evaluate to a `bool` `true` otherwise an alert is triggered.

The object that is passed as the root is the `integer: totalHits`.

IE:

```
  "check": "{{ eq . 0 }}"
```

Will check if `totalHits` is equal to `0`. If not, an alert gets thrown.

#### Body

This is the body of the alert. Once again, it's using the [go template](https://godoc.org/text/template) so you have a lot of flexibility.

The root of the object of type `AlertMetadata`:

```
type AlertMetadata struct {
  Query string
  File string
  TotalHits int64
  Name string
  Owners []string
}
```

IE:

```
  "body": "Host: localhost had \"{{ .TotalHits }}\" in the past minute"
```

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
    "owners": [string]
  }
}
```

Here is an example of a simple `shell` script that uses `curl` and `jq` to send the alert to `slack`:

```
#!/bin/sh

payload=$(cat -)
export payload

body="$(printenv payload | jq -jr '"<@" + .metadata.owners[] + "> " , .body')"
export body

curl -X POST -H 'Content-type: application/json' --data "{\"text\":'${body}'}" https://hooks.slack.com/services/SECRET
```

In this case, `Owners` will be `slack` handles. So if we pass `-owners blih,blah`, the output on `slack` will be like this:

```
@blih @blah Suspicious change on host: "localhost" from auditbeat
```

You are more than welcomed to re-use that script. In following versions I will try to include them directly to `esalert`.

## Usage

```
Usage of ./esalert: <-server=Url> <-index=Index> <-dir=Directory>
  -dir string
      Directory where the .json files are
  -exec string
      Execute a command when alerting
  -index string
      Specify the elasticsearch index to query
  -owners string
      List of default owners separated by "," to notify
  -server string
      Specify elasticsearch server to query (default "http://localhost:9200")
  -sleep-for int
      Sleep for in seconds after all queries have been ran (default 60)
```

#### Contributing

Issues, PRs and ideas are more than welcome. File one and I'll review it ASAP.
