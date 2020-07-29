package main

import (
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger_config "github.com/uber/jaeger-client-go/config"
	jaeger_metrics "github.com/uber/jaeger-lib/metrics/prometheus"
)

func main() {
	initJaeger("tracing example")

	server := instrumentedServer(handler)
	server.ListenAndServe()
}

func handler(w http.ResponseWriter, r *http.Request) {

}

func instrumentedServer(handler http.HandlerFunc) *http.Server {
	tracingMiddleware := func(w http.ResponseWriter, r *http.Request) {
		span := opentracing.SpanFromContext(r.Context())
		span.SetOperationName("Incoming HTTP Request")

		handler(w, r)

		span.Finish()
	}

	return &http.Server{
		Handler: http.HandlerFunc(tracingMiddleware),
		Addr:    "0.0.0.0:8000",
	}
}

// initJaeger returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func initJaeger(service string) {
	cfg, err := jaeger_config.FromEnv()
	if err != nil {
		panic(err)
	}
	metricsFactory := jaeger_metrics.New()
	cfg.InitGlobalTracer(service, jaeger_config.Metrics(metricsFactory))
}
