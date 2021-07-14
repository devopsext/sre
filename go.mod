module github.com/devopsext/sre

go 1.16

require (
	github.com/DataDog/datadog-go v4.7.0+incompatible
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/devopsext/utils v0.0.3
	github.com/google/uuid v1.2.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/xid v1.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	go.opentelemetry.io/otel v1.0.0-RC1
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0-RC1
	go.opentelemetry.io/otel/metric v0.21.0
	go.opentelemetry.io/otel/sdk v1.0.0-RC1
	go.opentelemetry.io/otel/sdk/metric v0.21.0
	go.opentelemetry.io/otel/trace v1.0.0-RC1
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/time v0.0.0-20210608053304-ed9ce3a009e4 // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.31.1
)

replace gopkg.in/DataDog/dd-trace-go.v1 => github.com/devopsext/dd-trace-go v1.31.2
