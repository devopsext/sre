package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"sync"
	"syscall"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/sre/provider"
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
}

var datadogOptions = provider.DataDogOptions{
	ServiceName: "",
	Environment: "none",
	Tags:        "",
}

var datadogTracerOptions = provider.DataDogTracerOptions{
	Host: "",
	Port: 8126,
}

var datadogLoggerOptions = provider.DataDogLoggerOptions{
	Host:  "",
	Port:  10518,
	Level: "info",
}

var datadogMetricerOptions = provider.DataDogMetricerOptions{
	Host:   "",
	Port:   10518,
	Prefix: "sre",
}

func interceptSyscall() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
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
			if common.HasElem(rootOptions.Logs, "stdout") {
				logs.Register(stdout)
			}

			datadogLoggerOptions.Version = VERSION
			datadogLoggerOptions.ServiceName = datadogOptions.ServiceName
			datadogLoggerOptions.Environment = datadogOptions.Environment
			datadogLoggerOptions.Tags = datadogOptions.Tags

			datadogLogger := provider.NewDataDogLogger(datadogLoggerOptions, logs, stdout)
			if common.HasElem(rootOptions.Logs, "datadog") {
				logs.Register(datadogLogger)
			}

			logs.Info("Booting...")

			// Metrics

			prometheusOptions.Version = VERSION
			prometheus := provider.NewPrometheus(prometheusOptions, logs, stdout)
			prometheus.SetCallerOffset(1)
			if common.HasElem(rootOptions.Metrics, "prometheus") {
				prometheus.Start(&mainWG)
				metrics.Register(prometheus)
			}

			datadogMetricerOptions.Version = VERSION
			datadogMetricerOptions.ServiceName = datadogOptions.ServiceName
			datadogMetricerOptions.Environment = datadogOptions.Environment
			datadogMetricerOptions.Tags = datadogOptions.Tags

			datadogMetricer := provider.NewDataDogMetricer(datadogMetricerOptions, logs, stdout)
			datadogMetricer.SetCallerOffset(1)
			if common.HasElem(rootOptions.Metrics, "datadog") {
				metrics.Register(datadogMetricer)
			}

			// Tracing

			jaegerOptions.Version = VERSION
			jaeger := provider.NewJaeger(jaegerOptions, logs, stdout)
			jaeger.SetCallerOffset(1)
			if common.HasElem(rootOptions.Traces, "jaeger") {
				traces.Register(jaeger)
			}

			datadogTracerOptions.Version = VERSION
			datadogTracerOptions.ServiceName = datadogOptions.ServiceName
			datadogTracerOptions.Environment = datadogOptions.Environment
			datadogTracerOptions.Tags = datadogOptions.Tags

			datadogTracer := provider.NewDataDogTracer(datadogTracerOptions, logs, stdout)
			datadogTracer.SetCallerOffset(1)
			if common.HasElem(rootOptions.Traces, "datadog") {
				traces.Register(datadogTracer)
			}

		},
		Run: func(cmd *cobra.Command, args []string) {

			logs.Info("Log message to every log provider...")

			rootSpan := traces.StartSpan()
			span := rootSpan

			logs.SpanInfo(rootSpan, "This message has correlation with span...")

			counter := metrics.Counter("calls", "Calls counter", []string{"label"})

			for i := 0; i < 10; i++ {

				span := traces.StartChildSpan(span.GetContext())
				span.SetName(fmt.Sprintf("sre-name-%d", i))

				time.Sleep(time.Duration(100*i) * time.Millisecond)
				counter.Inc(strconv.Itoa(i))
				logs.SpanDebug(span, "Counter increment %d", i)

				span.Finish()
			}

			logs.Info("Wait until it will be interrupted...")

			rootSpan.Finish()

			mainWG.Wait()
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringSliceVar(&rootOptions.Logs, "logs", rootOptions.Logs, "Log providers: stdout, datadog")
	flags.StringSliceVar(&rootOptions.Metrics, "metrics", rootOptions.Metrics, "Metric providers: prometheus, datadog")
	flags.StringSliceVar(&rootOptions.Traces, "traces", rootOptions.Traces, "Trace providers: jaeger, datadog")

	flags.StringVar(&stdoutOptions.Format, "stdout-format", stdoutOptions.Format, "Stdout format: json, text, template")
	flags.StringVar(&stdoutOptions.Level, "stdout-level", stdoutOptions.Level, "Stdout level: info, warn, error, debug, panic")
	flags.StringVar(&stdoutOptions.Template, "stdout-template", stdoutOptions.Template, "Stdout template")
	flags.StringVar(&stdoutOptions.TimestampFormat, "stdout-timestamp-format", stdoutOptions.TimestampFormat, "Stdout timestamp format")
	flags.BoolVar(&stdoutOptions.TextColors, "stdout-text-colors", stdoutOptions.TextColors, "Stdout text colors")

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

	flags.StringVar(&datadogOptions.ServiceName, "datadog-service-name", datadogOptions.ServiceName, "DataDog service name")
	flags.StringVar(&datadogOptions.Environment, "datadog-environment", datadogOptions.Environment, "DataDog environment")
	flags.StringVar(&datadogOptions.Tags, "datadog-tags", datadogOptions.Tags, "DataDog tags")

	flags.StringVar(&datadogTracerOptions.Host, "datadog-tracer-host", datadogTracerOptions.Host, "DataDog tracer host")
	flags.IntVar(&datadogTracerOptions.Port, "datadog-tracer-port", datadogTracerOptions.Port, "Datadog tracer port")

	flags.StringVar(&datadogLoggerOptions.Host, "datadog-logger-host", datadogLoggerOptions.Host, "DataDog logger host")
	flags.IntVar(&datadogLoggerOptions.Port, "datadog-logger-port", datadogLoggerOptions.Port, "Datadog logger port")
	flags.StringVar(&datadogLoggerOptions.Level, "datadog-logger-level", datadogLoggerOptions.Level, "DataDog logger level: info, warn, error, debug, panic")

	flags.StringVar(&datadogMetricerOptions.Host, "datadog-metricer-host", datadogMetricerOptions.Host, "DataDog metricer host")
	flags.IntVar(&datadogMetricerOptions.Port, "datadog-metricer-port", datadogMetricerOptions.Port, "Datadog metricer port")
	flags.StringVar(&datadogMetricerOptions.Prefix, "datadog-metricer-prefix", datadogMetricerOptions.Prefix, "DataDog metricer prefix")

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
