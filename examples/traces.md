# Traces example

## Run example

```sh
go run traces.go
```

Output should be:
```log
...
sre/provider/jaeger.go:459 Jaeger tracer is up...
sre/examples/traces.go:66 DataDog tracer is disabled.
sre/provider/opentelemetry.go:451 Opentelemetry tracer is up...
sre/provider/opentelemetry.go:250 hex encoded trace-id must have length equals to 32
sre/examples/traces.go:24 Counter increment 0
sre/examples/traces.go:24 Counter increment 1
sre/examples/traces.go:24 Counter increment 2
sre/examples/traces.go:24 Counter increment 3
sre/examples/traces.go:24 Counter increment 4
sre/examples/traces.go:24 Counter increment 5
sre/examples/traces.go:24 Counter increment 6
sre/examples/traces.go:24 Counter increment 7
sre/examples/traces.go:24 Counter increment 8
```

## Go to Jaeger UI and there should be seen

![Jaeger](/jaeger.png)

## For a case of Opentelemetry (Lightstep) screenshoot will be

![Lightstep](/lightstep.png)
