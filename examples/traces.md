# Traces example

## Run example

```sh
go run traces.go
```

Output should be:
```log
...
sre@v0.0.6/provider/jaeger.go:444 Jaeger tracer is up...
go/sre/traces.go:66 DataDog tracer is disabled.
sre@v0.0.6/provider/opentelemetry.go:493 Opentelemetry tracer is up...
go/sre/traces.go:24 Counter increment 0
go/sre/traces.go:24 Counter increment 1
go/sre/traces.go:24 Counter increment 2
...
```

## Go to Jaeger UI and there should be seen

![Jaeger](/jaeger.png)

## For a case of Opentelemetry (Lightstep) screenshoot will be

![Lightstep](/lightstep.png)
