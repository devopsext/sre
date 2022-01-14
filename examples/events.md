# Events example

## Run example

```sh
go run events.go
```

Output should be:
```log
...
sre/provider/grafana.go:153 Grafana eventer is up...
sre/provider/newrelic.go:872 NewRelic eventer is up...
sre/provider/datadog.go:796 DataDog eventer is up...
sre/provider/grafana.go:127 Annotation 1502806. Annotation added
sre/provider/grafana.go:127 Annotation 1502807. Annotation added
sre/provider/grafana.go:127 Annotation 1502808. Annotation added
```

## Grafana UI

![Grafana](/examples/grafana-events.jpg)

## NewRelic UI

![NewRelic](/examples/newrelic-events.jpg)

## DataDog UO
![DataDog](/examples/datadog-events.png)
![DataDog](/examples/datadog-events2.png)