# Metrics example

## Run example

```sh
go run metrics.go
```

Output should be:
```log
sre/provider/prometheus.go:78 Start prometheus endpoint...
sre/provider/prometheus.go:88 Prometheus is up. Listening...
sre/provider/datadog.go:665 DataDog meter is up...
```

## Prometheus endpoint

```sh
curl -sk http://127.0.0.1:8080/metrics | grep sre_
```
```prometheus
# HELP sre_calls Calls counter
# TYPE sre_calls counter
sre_calls{time="2021-06-17 18:00:30.248990729 +0300 EEST m=+0.002878298"} 1
```

## NewRelic UI

![NewRelic](/examples/newrelic-metrics.png)
