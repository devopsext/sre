module github.com/devopsext/sre

go 1.24

require (
	github.com/DataDog/datadog-api-client-go v1.7.0
	github.com/DataDog/datadog-go v4.7.0+incompatible
	github.com/VictoriaMetrics/metrics v1.40.0
	github.com/devopsext/utils v0.4.0
	github.com/newrelic/newrelic-telemetry-sdk-go v0.8.1
	github.com/opentracing/opentracing-go v1.2.0
	github.com/rs/xid v1.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.4.0
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	gopkg.in/DataDog/dd-trace-go.v1 v1.31.1
)

require (
	github.com/DataDog/sketches-go v1.0.0 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/go-logr/logr v1.2.1 // indirect
	github.com/go-logr/stdr v1.2.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/tinylib/msgp v1.1.2 // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/valyala/fastrand v1.1.0 // indirect
	github.com/valyala/histogram v1.2.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/time v0.0.0-20210608053304-ed9ce3a009e4 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.1 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.42.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace gopkg.in/DataDog/dd-trace-go.v1 => github.com/devopsext/dd-trace-go v1.31.2

//replace	github.com/newrelic/newrelic-telemetry-sdk-go => github.com/devopsext/newrelic-telemetry-sdk-go v0.8.2
