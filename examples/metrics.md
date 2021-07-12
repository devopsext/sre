# Metrics example

## Run example

```sh
go run metrics.go
```
```log
sre@v0.0.6/provider/prometheus.go:83 Start prometheus endpoint...
sre@v0.0.6/provider/prometheus.go:93 Prometheus is up. Listening...
sre@v0.0.6/provider/datadog.go:648 DataDog meter is up...
```

## Check Prometheus endpoint

```sh
curl -sk http://127.0.0.1:8080/metrics | grep sre_
```
```prometheus
# HELP sre_calls Calls counter
# TYPE sre_calls counter
sre_calls{time="2021-06-17 18:00:30.248990729 +0300 EEST m=+0.002878298"} 1
```
