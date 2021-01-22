module github.com/joe-elliott/tracing-example

go 1.14

require (
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/go-kit/kit v0.9.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.opentelemetry.io/contrib/instrumentation/net/http v0.11.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.16.0
	go.opentelemetry.io/otel v0.16.0
	go.opentelemetry.io/otel/exporters/otlp v0.16.0
	go.opentelemetry.io/otel/sdk v0.16.0
	go.uber.org/atomic v1.6.0 // indirect
	google.golang.org/grpc v1.34.0
)
