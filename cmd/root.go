package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"sync"
	"syscall"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/sre/provider"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var VERSION = "unknown"

var logs = common.NewLogs()
var traces = common.NewTraces()
var metrics = common.NewMetrics()
var stdout *provider.Stdout
var mainWG sync.WaitGroup

type RootOptions struct {
	Logs    []string
	Metrics []string
	Traces  []string
}

var rootOptions = RootOptions{

	Logs:    []string{"stdout"},
	Metrics: []string{"prometheus"},
	Traces:  []string{},
}

var stdoutOptions = provider.StdoutOptions{

	Format:          "text",
	Level:           "info",
	Template:        "{{.file}} {{.msg}}",
	TimestampFormat: time.RFC3339Nano,
	TextColors:      true,
	Debug:           false,
}

var prometheusOptions = provider.PrometheusOptions{

	URL:    "/metrics",
	Listen: "127.0.0.1:8080",
	Prefix: "sre",
}

var jaegerOptions = provider.JaegerOptions{
	ServiceName:         "sre",
	AgentHost:           "",
	AgentPort:           6831,
	Endpoint:            "",
	User:                "",
	Password:            "",
	BufferFlushInterval: 0,
	QueueSize:           0,
	Tags:                "",
	Debug:               false,
}

var datadogOptions = provider.DataDogOptions{
	ServiceName: "",
	Environment: "none",
	Tags:        "",
	Debug:       false,
}

var datadogTracerOptions = provider.DataDogTracerOptions{
	AgentHost: "",
	AgentPort: 8126,
}

var datadogLoggerOptions = provider.DataDogLoggerOptions{
	AgentHost: "",
	AgentPort: 10518,
	Level:     "info",
}

var datadogMeterOptions = provider.DataDogMeterOptions{
	AgentHost: "",
	AgentPort: 8125,
	Prefix:    "sre",
}

var opentelemetryOptions = provider.OpentelemetryOptions{
	ServiceName: "",
	Environment: "",
	Attributes:  "",
	Debug:       false,
}

var opentelemetryTracerOptions = provider.OpentelemetryTracerOptions{
	AgentHost: "",
	AgentPort: 4317,
}

var opentelemetryMeterOptions = provider.OpentelemetryMeterOptions{
	AgentHost:     "",
	AgentPort:     4317,
	Prefix:        "sre",
	CollectPeriod: 1000,
}

var newrelicOptions = provider.NewRelicOptions{
	ServiceName: "",
	Environment: "",
	Labels:      "",
	Debug:       false,
}

var newrelicLoggerOptions = provider.NewRelicLoggerOptions{
	AgentHost: "",
	AgentPort: 5171,
	Level:     "info",
}

func interceptSyscall() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		logs.Info("Exiting...")
		os.Exit(1)
	}()
}

func Execute() {

	rootCmd := &cobra.Command{
		Use:   "sre",
		Short: "SRE",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {

			stdoutOptions.Version = VERSION
			stdout = provider.NewStdout(stdoutOptions)
			stdout.SetCallerOffset(2)
			if utils.Contains(rootOptions.Logs, "stdout") {
				logs.Register(stdout)
			}

			datadogLoggerOptions.Version = VERSION
			datadogLoggerOptions.ServiceName = datadogOptions.ServiceName
			datadogLoggerOptions.Environment = datadogOptions.Environment
			datadogLoggerOptions.Tags = datadogOptions.Tags
			datadogLogger := provider.NewDataDogLogger(datadogLoggerOptions, logs, stdout)
			if utils.Contains(rootOptions.Logs, "datadog") && datadogLogger != nil {
				logs.Register(datadogLogger)
			}

			newrelicLoggerOptions.Version = VERSION
			newrelicLoggerOptions.ServiceName = newrelicOptions.ServiceName
			newrelicLoggerOptions.Environment = newrelicOptions.Environment
			newrelicLoggerOptions.Labels = newrelicOptions.Labels
			newrelicLogger := provider.NewNewRelicLogger(newrelicLoggerOptions, logs, stdout)
			if utils.Contains(rootOptions.Logs, "newrelic") && newrelicLogger != nil {
				logs.Register(newrelicLogger)
			}

			logs.Info("Booting...")

			// Metrics

			prometheusOptions.Version = VERSION
			prometheus := provider.NewPrometheusMeter(prometheusOptions, logs, stdout)
			if utils.Contains(rootOptions.Metrics, "prometheus") && prometheus != nil {
				prometheus.StartInWaitGroup(&mainWG)
				metrics.Register(prometheus)
			}

			datadogMeterOptions.Version = VERSION
			datadogMeterOptions.ServiceName = datadogOptions.ServiceName
			datadogMeterOptions.Environment = datadogOptions.Environment
			datadogMeterOptions.Tags = datadogOptions.Tags
			datadogMeter := provider.NewDataDogMeter(datadogMeterOptions, logs, stdout)
			if utils.Contains(rootOptions.Metrics, "datadog") && datadogMeter != nil {
				metrics.Register(datadogMeter)
			}

			opentelemetryMeterOptions.Version = VERSION
			opentelemetryMeterOptions.ServiceName = opentelemetryOptions.ServiceName
			opentelemetryMeterOptions.Environment = opentelemetryOptions.Environment
			opentelemetryMeterOptions.Attributes = opentelemetryOptions.Attributes
			opentelemetryMeter := provider.NewOpentelemetryMeter(opentelemetryMeterOptions, logs, stdout)
			if utils.Contains(rootOptions.Metrics, "opentelemetry") && opentelemetryMeter != nil {
				metrics.Register(opentelemetryMeter)
			}

			// Tracing

			jaegerOptions.Version = VERSION
			jaeger := provider.NewJaegerTracer(jaegerOptions, logs, stdout)
			if utils.Contains(rootOptions.Traces, "jaeger") && jaeger != nil {
				traces.Register(jaeger)
			}

			datadogTracerOptions.Version = VERSION
			datadogTracerOptions.ServiceName = datadogOptions.ServiceName
			datadogTracerOptions.Environment = datadogOptions.Environment
			datadogTracerOptions.Tags = datadogOptions.Tags
			datadogLoggerOptions.Debug = datadogOptions.Debug
			datadogTracer := provider.NewDataDogTracer(datadogTracerOptions, logs, stdout)
			if utils.Contains(rootOptions.Traces, "datadog") && datadogTracer != nil {
				traces.Register(datadogTracer)
			}

			opentelemetryTracerOptions.Version = VERSION
			opentelemetryTracerOptions.ServiceName = opentelemetryOptions.ServiceName
			opentelemetryTracerOptions.Environment = opentelemetryOptions.Environment
			opentelemetryTracerOptions.Attributes = opentelemetryOptions.Attributes
			opentelemtryTracer := provider.NewOpentelemetryTracer(opentelemetryTracerOptions, logs, stdout)
			if utils.Contains(rootOptions.Traces, "opentelemetry") && opentelemtryTracer != nil {
				traces.Register(opentelemtryTracer)
			}

		},
		Run: func(cmd *cobra.Command, args []string) {

			defer metrics.Stop()
			defer traces.Stop()

			logs.Info("Log message to every log provider...")

			rootSpan := traces.StartSpan()
			rootSpan.SetBaggageItem("some-restriction", "enabled")
			spanCtx := rootSpan.GetContext()
			if spanCtx != nil {
				logs.Info("Trace ID is %s", spanCtx.GetTraceID())
				logs.Info("Span ID is %s", spanCtx.GetSpanID())
			}

			logs.SpanInfo(rootSpan, "This message has correlation with span...")

			counter := metrics.Counter("calls", "Calls counter", []string{"label"}, "counter", "of", "iteration")

			for i := 0; i < 10; i++ {

				span := traces.StartChildSpan(spanCtx)
				span.SetName(fmt.Sprintf("name-%d", i))

				time.Sleep(time.Duration(100*i) * time.Millisecond)
				counter.Inc(strconv.Itoa(i))
				logs.SpanDebug(span, "Counter increment %d", i)

				spanCtx = span.GetContext()
				span.Finish()
			}

			span := traces.StartChildSpan(rootSpan.GetContext())
			span.SetName("call")

			content, err := ioutil.ReadFile("k8s.json")
			if err != nil {
				logs.SpanError(span, err)
			}
			reader := bytes.NewReader(content)

			req, err := http.NewRequest("POST", "http://127.0.0.1:8081/k8s", reader)
			if err != nil {
				logs.SpanError(span, err)
			}

			req.Header.Set("Content-Type", "application/json")
			ctx := span.GetContext()
			if ctx == nil {
				logs.SpanError(span, "no span context found")
				span.Finish()
				rootSpan.Finish()
				traces.Stop()
				os.Exit(0)
			}
			traceID := ctx.GetTraceID()
			req.Header.Set("X-Trace-ID", traceID)

			client := common.MakeHttpClient(5000)

			resp, err := client.Do(req)
			if err != nil {
				logs.SpanError(span, err)
				span.Finish()
				rootSpan.Finish()
				traces.Stop()
				os.Exit(0)
			}

			defer resp.Body.Close()

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logs.SpanError(span, err)
			}
			logs.Info(b)

			span.Finish()
			logs.Info("Wait until it will be interrupted...")

			rootSpan.Finish()
			mainWG.Wait()
			traces.Stop()
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringSliceVar(&rootOptions.Logs, "logs", rootOptions.Logs, "Log providers: stdout, datadog, newrelic")
	flags.StringSliceVar(&rootOptions.Metrics, "metrics", rootOptions.Metrics, "Metric providers: prometheus, datadog, opentelemetry")
	flags.StringSliceVar(&rootOptions.Traces, "traces", rootOptions.Traces, "Trace providers: jaeger, datadog, opentelemetry")

	flags.StringVar(&stdoutOptions.Format, "stdout-format", stdoutOptions.Format, "Stdout format: json, text, template")
	flags.StringVar(&stdoutOptions.Level, "stdout-level", stdoutOptions.Level, "Stdout level: info, warn, error, debug, panic")
	flags.StringVar(&stdoutOptions.Template, "stdout-template", stdoutOptions.Template, "Stdout template")
	flags.StringVar(&stdoutOptions.TimestampFormat, "stdout-timestamp-format", stdoutOptions.TimestampFormat, "Stdout timestamp format")
	flags.BoolVar(&stdoutOptions.TextColors, "stdout-text-colors", stdoutOptions.TextColors, "Stdout text colors")
	flags.BoolVar(&stdoutOptions.Debug, "stdout-debug", stdoutOptions.Debug, "Stdout debug")

	flags.StringVar(&prometheusOptions.URL, "prometheus-url", prometheusOptions.URL, "Prometheus endpoint url")
	flags.StringVar(&prometheusOptions.Listen, "prometheus-listen", prometheusOptions.Listen, "Prometheus listen")
	flags.StringVar(&prometheusOptions.Prefix, "prometheus-prefix", prometheusOptions.Prefix, "Prometheus prefix")

	flags.StringVar(&jaegerOptions.ServiceName, "jaeger-service-name", jaegerOptions.ServiceName, "Jaeger service name")
	flags.StringVar(&jaegerOptions.AgentHost, "jaeger-agent-host", jaegerOptions.AgentHost, "Jaeger agent host")
	flags.IntVar(&jaegerOptions.AgentPort, "jaeger-agent-port", jaegerOptions.AgentPort, "Jaeger agent port")
	flags.StringVar(&jaegerOptions.Endpoint, "jaeger-endpoint", jaegerOptions.Endpoint, "Jaeger endpoint")
	flags.StringVar(&jaegerOptions.User, "jaeger-user", jaegerOptions.User, "Jaeger user")
	flags.StringVar(&jaegerOptions.Password, "jaeger-password", jaegerOptions.Password, "Jaeger password")
	flags.IntVar(&jaegerOptions.BufferFlushInterval, "jaeger-buffer-flush-interval", jaegerOptions.BufferFlushInterval, "Jaeger buffer flush interval")
	flags.IntVar(&jaegerOptions.QueueSize, "jaeger-queue-size", jaegerOptions.QueueSize, "Jaeger queue size")
	flags.StringVar(&jaegerOptions.Tags, "jaeger-tags", jaegerOptions.Tags, "Jaeger tags, comma separated list of name=value")
	flags.BoolVar(&jaegerOptions.Debug, "jaeger-debug", jaegerOptions.Debug, "Jaeger debug")

	flags.StringVar(&datadogOptions.ServiceName, "datadog-service-name", datadogOptions.ServiceName, "DataDog service name")
	flags.StringVar(&datadogOptions.Environment, "datadog-environment", datadogOptions.Environment, "DataDog environment")
	flags.StringVar(&datadogOptions.Tags, "datadog-tags", datadogOptions.Tags, "DataDog tags")
	flags.BoolVar(&datadogOptions.Debug, "datadog-debug", datadogOptions.Debug, "DataDog debug")
	flags.StringVar(&datadogTracerOptions.AgentHost, "datadog-tracer-agent-host", datadogTracerOptions.AgentHost, "DataDog tracer agent host")
	flags.IntVar(&datadogTracerOptions.AgentPort, "datadog-tracer-agent-port", datadogTracerOptions.AgentPort, "Datadog tracer agent port")
	flags.StringVar(&datadogLoggerOptions.AgentHost, "datadog-logger-agent-host", datadogLoggerOptions.AgentHost, "DataDog logger agent host")
	flags.IntVar(&datadogLoggerOptions.AgentPort, "datadog-logger-agent-port", datadogLoggerOptions.AgentPort, "Datadog logger agent port")
	flags.StringVar(&datadogLoggerOptions.Level, "datadog-logger-level", datadogLoggerOptions.Level, "DataDog logger level: info, warn, error, debug, panic")
	flags.StringVar(&datadogMeterOptions.AgentHost, "datadog-meter-agent-host", datadogMeterOptions.AgentHost, "DataDog meter agent host")
	flags.IntVar(&datadogMeterOptions.AgentPort, "datadog-meter-agent-port", datadogMeterOptions.AgentPort, "Datadog meter agent port")
	flags.StringVar(&datadogMeterOptions.Prefix, "datadog-meter-prefix", datadogMeterOptions.Prefix, "DataDog meter prefix")

	flags.StringVar(&opentelemetryOptions.ServiceName, "opentelemetry-service-name", opentelemetryOptions.ServiceName, "Opentelemetry service name")
	flags.StringVar(&opentelemetryOptions.Environment, "opentelemetry-environment", opentelemetryOptions.Environment, "Opentelemetry environment")
	flags.StringVar(&opentelemetryOptions.Attributes, "opentelemetry-attributes", opentelemetryOptions.Attributes, "Opentelemetry attributes")
	flags.StringVar(&opentelemetryTracerOptions.AgentHost, "opentelemetry-tracer-agent-host", opentelemetryTracerOptions.AgentHost, "Opentelemetry tracer agent host")
	flags.IntVar(&opentelemetryTracerOptions.AgentPort, "opentelemetry-tracer-agent-port", opentelemetryTracerOptions.AgentPort, "Opentelemetry tracer agent port")
	flags.StringVar(&opentelemetryMeterOptions.AgentHost, "opentelemetry-meter-agent-host", opentelemetryMeterOptions.AgentHost, "Opentelemetry meter agent host")
	flags.IntVar(&opentelemetryMeterOptions.AgentPort, "opentelemetry-meter-agent-port", opentelemetryMeterOptions.AgentPort, "Opentelemetry meter agent port")
	flags.StringVar(&opentelemetryMeterOptions.Prefix, "opentelemetry-meter-prefix", opentelemetryMeterOptions.Prefix, "Opentelemetry meter prefix")

	flags.StringVar(&newrelicOptions.ServiceName, "newrelic-service-name", newrelicOptions.ServiceName, "NewRelic service name")
	flags.StringVar(&newrelicOptions.Environment, "newrelic-environment", newrelicOptions.Environment, "NewRelic environment")
	flags.StringVar(&newrelicOptions.Labels, "newrelic-labels", newrelicOptions.Labels, "NewRelic labels")
	flags.StringVar(&newrelicLoggerOptions.AgentHost, "newrelic-logger-agent-host", newrelicLoggerOptions.AgentHost, "NewRelic logger agent host")
	flags.IntVar(&newrelicLoggerOptions.AgentPort, "newrelic-logger-agent-port", newrelicLoggerOptions.AgentPort, "NewRelic logger agent port")
	flags.StringVar(&newrelicLoggerOptions.Level, "newrelic-logger-level", newrelicLoggerOptions.Level, "NewRelic logger level: info, warn, error, debug, panic")

	interceptSyscall()

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(VERSION)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		logs.Error(err)
		os.Exit(1)
	}
}
