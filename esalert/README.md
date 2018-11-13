# Esalert

Create powerfull rules based on [elasticsearch](https://www.elastic.co/products/elasticsearch), stay alerted and trigger active-response to have complete control of your infrastructure.

Work less, work smarter.

Features:

  - Really fast: written in Go with multi-threading query support
  - Active-response: automate and deploy responses based on alerts to your infrastructure
  - Ultra customizable: write your own scripts, templates, customize your rules that fits exactly to your need

## Design

`Esalert` has two components:

  - a server: `esalertd`
  - and a client: `eslert` (optional)

### Esalertd

It is the core of the system. For now it's a single process which processes a number of rules (json files) then schedule them to query elasticsearch based on a date range.

The date range is determined by some date arithmetic from the Go [time](https://godoc.org/time) library.

Once the query has ran, it gets either the `total_hits` or the number of buckets from an `aggregation`. That number is passed through a template which will be evaluated as boolan: `true` means that the check of the rules holds true, no need to alert. If it is not true an alert will be triggered.

When an alert is triggered, it executes the binary either from the rule or from the command line flags. A `json` object is passed to the `stdin` of the program, the logic is to be implemented by the user. An alert can only be triggered once per rule and per time range. It won't alert more than once per minute if the queried range is one minute, even if the rule is scheduled every second.

As soon as the alert is triggered, a scroll is scheduled on the exact same query and saved in memory so you have logs of what happened. The scroll happens asynchronously so it does not interfere with the rule/alert scheduler.

Finally, if configured, when alerts are thrown, an `active-response` can be triggered. Active-responses are orders to the `esalert` client side to execute a command with some parameters. The parameters are derived from the result of the query. They also have an expire date which is tells the client to revert the order (let's say you want to ban an ip address for 2 minutes if it behaved badly).

### Esalert

The client side will simple query the server for actions to execute.

Every 5 seconds, the client will query the server for all the responses for all the tags the client is handeling. It will merge the responses to be triggered once if they are from multiple different tags.

Note that it is possible an alert is triggered multiple times and the same value before the expiry date. Active-response scripts must handle actions like a stack: add/pop:

  - GOOD: add x times the same iptables rule which might ban the same ip address regardless of the expiry time, remove one bad one like you don't know the global state.
  - BAD: remove all iptables rules when a stop is called.

## Active-Response

This is the exciting new feature. Every time an alert is thrown, an active-response per tags gets generated. The active-response has an expiry time that enables you to revert the action when the time has expires.

The client/server architecture is actually a fetch mode rather than pull mode. The server stores the active-responses and clients periodically fetches what they need to fetch.

There is no encryption/authentication so *please* encrypt your traffic or wait until I come up with a solution. Use at your own risk or contact me.

### Tags

Tags are hardcoded values in the rule's configuration that will indicate where the generated active-response will be.

Let's say you have 3 different kind of rules:

  - A rule that will trigger an active-response on all your datacenter `us-east-1`
  - A rule that will trigger an active-response on all your backend servers `backend`
  - A rule that will trigger an active-response on all your servers `all`

Then you configure your client to "listen" on all the tags: `--tags all --tags backend -- tags us-east-1`.

Next time an active-response gets triggered, the clients that are listening on the right tags will be triggered.

### IDs

Active-responses have `id`s. As multiple tags can be specified in the rule's configuration, they will be duplicated for each tags. Clients listening on multiple tags can trigger the same active-response. To mitigate that, the client will only trigger active-responses that have not been seen before.

### Expiry

You can specify for how long the action-response will be active. When it is active, the client invoke the action with the keyword "start" and some parameter. When the active-response expires, the clients must invoke the same script and the same parameters, but with the keyword "stop" instead.

## Rules

They are at the core of `esalertd`'s configuration. Each rule is a json file with a lot of fields inside. For convenience, you can generate your own using a simple shell script and a template.

Here are the options:

  - `query`: `string`
    - Elasticsearch query using a go template. It uses the object `TemplateQueryRoot` as root of the template.

  - `body`: `string`
    - Go template of the body of the alert. It uses the object `AlertPayload` as root of the template.
  - `check`: `string`
    - This is the most important field, if it returns anything except "true", an alert is thrown. It uses the number of results found from the elasticsearch query as root of the template.
  - `log`: `string`
    - Go template of the alert logs. When the alert is thrown and a job is schedule to scroll through all the results, you can specify a template so you can customize the output. Note that the result is not html safe, the content-type from the http header is `application/text` to avoid any log poisoning attack. The root of the template is the object returned by elasticsearch.
  - `alert_every`: `duration`
    - Trigger an alert no less than the specified duration. It is useful if you query often but for a very large range, you might not want to be bothered that much. Note that only the first alert is recorded. If further alerts are triggering, it will increment a counter and get logged, but it will not be triggered.
  - `max_wait_schedule`: `duration`
    - Every second the scheduler runs and tries to find queries that are due to be scheduled. However, if all the queries are due to be scheduled at the same time, it will lead to a lot of traffic every x period of time. To mitigate this, you can use this setting, it will randomly wait between 0ns and this duration to first schedule the query.
  - `run_every`: `duration`
    - Run the query every x period of time. Note that it won't be really accurate (because of clocks and OS schedulers). Be careful tunning this setting, a date range smaller than this setting might make you skip data (some times you don't care about skipping data).
  - `owners`: `stringarray`
    - Set the owners of this rules. The owners are passed to the alert script so you can contact the right people
  - `aggregation`: `object`
    - Root object field when you need to aggregate on some fields. If not specified, esalertd will use the `total_hits` as value.
  - `aggregation.type`: `string`
    - internal aggregation type to esalertd. It does not reflect the elasticsearch types as more logic is needed.
  - `aggregation.field`: `string`
    - set the field on which to perform the aggregation
  - `from`: `object`
    - the `from` field is a `DateTime` internal object that sets the *greater or equal than* of the query range. It performs date arithmetics so the time stays consistent throughout the query lifecycle.
  - `from.date`: `string`
    - specify the date of the query range. It accepts the standard `json` standard format by default but you can specify a layout if you want to parse to a custom format. By default it is set to `now`.
  - `from.round`: `string`
    - round the date down to the nearest unit : `nanosecond`, `second`, `minute`, `day` and `week`. By default it rounds down to the minute
  - `from.minus`: `duration`
    - subtract the duration to the time. By default it is `60s` for the `from` and `0s` for the `to`.
  - `from.plus`: `duration`
    - add the duration to the time.
  - `from.layout`: `string`
    - specify the layout you want to use to parse the `date`. It uses the Go [time](https://godoc.org/time) package formatting.
  - `to`: `object`
    - same as the `from` object except sets the *lower than* of the query.
  - `response`: `object`
    - this it the `active-response` feature. When not specified, the `active-response` is not triggered
  - `response.expire`: `duration`
    - set the expiration date from the date that the alert has been triggered
  - `response.tags`: `string array`
    - set the associated tags with the `active-response`
  - `response.action`: `string`
    - command to execute on the client side
  - `response.args`: `string array`
    - list of arguments to pass to the script. Each argument is passed through the `TemplateResponseRoot` template.

## Aggregations

They are part of a the elasticsearch features. Some times you might not want just the total of all the hits. You might want to aggregate data over that query. An example would be to aggregate by ip address.

Then the `check` template will perform the check on each result from the aggregation. It will trigger alert *per aggregation result* and not overall.

Here is a list of all supported aggregation:

  - "terms": Simple `terms` aggregation over all the terms (nyi but the goal is to use partitioning to get all data in order to not miss a single one (slow))

## Templates

Esalert use extensively the Go [template](https://godoc.org/text/template) package.

In this section you will find multiple root object used in various template string.

### TemplateQueryRoot

  - `From`: string

    - date in `json` format of the from parameter from the running query.

  - `To`: string
    - data in `json` format of the to parameter from the running query.

### AlertPayload

  - `from`: `date`

    - the *greater or equal than* part of the query

  - `to`: `date`
    - the *lower than* part of the query
  - `scheduled_at`: `date`
    - when the query has been scheduled
  - `triggered_at`: `date`
    - when the alert has been triggered
  - `executed_at`: `date`
    - when the alert script has been executed
  - `count`: `int`
    - the result of the query
  - `value`: `string`
    - the value of the result. Will be the same as `count` if the query type is not an aggregation
  - `id`: `string`
    - a uniq id for the alert
  - `owners`: `string array`
    - a list of owners that need to be notified
  - `body`: `string`
    - the generated body of the alert
  - `rule_name`: `string`
    - the name of the rule that has gave birth to this alert
  - `log_url`: `string`
    - the url of where to get the log. Note that if the query is not an aggregation the `log_url` value will be empty
  - `metadata`: `array of strings`
    - specified metadata in the rule's configuration that are reflected to the alert payload
  - `alert`: `boolean`:
    - when `alert` is true, it means that the alert plugin should send an alert to notify people about the issue. It is synced with `alert_every`.

### TemplateResponseRoot

  - `Value`: `string|int`

    - Value from the query. It will the same as `count` if the query type is not an aggregation.

  - `Count`: `int`
    - Number of time the value has been seen. It will be the `total_hits` value if the query type is not an aggregation

## Caveats
no HA
memory only: logs, active-responses and so on
no authentication/encryption

## Date range/Run every/Query delay
All rules must have a date range. The date range is resolved when the rule is scheduled to be run. There is a lot of flexibility on how to configure the date range. Let's stat with the `date` field. This field can be any date format as long as you specify the layout in the `layout` field according to the go [time](https://godoc.org/time) package.

There is a special keyword `now` that you can use to set the time when the rule is scheduled. Then the time is rounded down to some unit. If you want precision, round it to the second. The larger the round is, the less it will be precise. 

## Rule examples

```
{
  "query": "filebeat.www_wp_nginx_access.http_code: 404",
        "check": "{{ eq . 0 }}",
  "body": "IP address: \"{{ .Value }}\" has hit the server more than 3 times on not found page: {{ .Count }}",
  "log": "{{ .message }}{{ newline }}",
  "aggregation": { 
    "type": "terms",
    "field": "filebeat.www_wp_nginx_access.remote_ip"
  },
  "alert_every": "10s",
  "max_wait_schedule": "10s",
  "owners": [
    "blah"
  ],
  "run_every": "1s",
  "from": {
    "date": "now",
    "round": "second",
    "minus": "15s"
  },
  "to": {
    "date": "now",
    "round": "second"
  },
  "response": {
    "expire": "30s",
    "tags": [
      "test",
      "test1"
    ],
    "action": "echo",
    "args": [
      "ban",
      "{{ .Value }}"
    ]
  }
}
```

This rule is big, but it outlines all the features of `esalertd`.

In essence it says:
  - Schedule this query not more than 10 sec after `esalertd` starts
  - Run this elasticsearch query: `filebeat.www_wp_nginx_access.http_code: 404` every second
  - From now minus 15 second rounded to the second
  - To now rounded to the second
  - Aggregate the results using the `terms` aggregation on the `filebeat.www_wp_nginx_access.remote_ip` field
  - Check if the number returned by each bucket is higher than `0`
  - Send an alert with user `blah` as owner if it does not pass the check test
  - Don't send more than 1 alert every 10 sec
  - Send the alert with the following template body: `IP address: "{{ .Value }}" has hit the server more than 3 times on not found page: {{ .Count }}`
  - Trigger a response that expires in 30 sec
  - With the following tags: `test` and `test1`
  - That will execute the following template command: `echo ban {{ .Value }}`

## Alert examples
## Active-response examples

## Contributions
You are more than welcome to contribute! Right now everything is a little bit messy so just open a PR or an issue and we can talk about it.

## Esalertd
### Usage
```
Usage of esalertd:
      --dir string             Directory where the .json files are
      --exec string            Execute a command when alerting
      --index string           Specify the elasticsearch index to query
      --listen string          Start HTTP server and listen in ip:port (default ":7769")
      --owners stringArray     List of default owners to notify
      --public-url string      Public facing URL (default "http://172.17.0.2:7769")
      --query-delay duration   When using "now", delay the query to allow index time (default 1s)
      --server string          Specify elasticsearch server to query (default "http://localhost:9200")
```

## Esalert
### Usage
```
Usage of esalert:
      --dir string         Path to actions scripts' directory
      --tags stringArray   Tags associated with this client
      --url string         Url of server
```
