package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"google.golang.org/grpc"

	"github.com/go-kit/kit/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

// global vars...gasp!
var addr = "127.0.0.1:8000"
var tracer trace.Tracer
var httpClient http.Client

func main() {
	flush := initTracer()
	defer flush()

	server := instrumentedServer(handler)

	fmt.Println("listening...")
	server.ListenAndServe()
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	longRunningProcess(ctx)

	// check cache
	if shouldExecute(40) {
		url := "http://" + addr + "/"

		resp, err := instrumentedGet(ctx, url)
		defer resp.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	// query database
	if shouldExecute(40) {
		url := "http://" + addr + "/"

		resp, err := instrumentedGet(ctx, url)
		defer resp.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func shouldExecute(percent int) bool {
	return rand.Int()%100 < percent
}

func longRunningProcess(ctx context.Context) {
	//span, _ := opentracing.StartSpanFromContext(ctx, "Long Running Process")
	//defer span.Finish()

	//span.SetTag("list length", 50)
	//time.Sleep(time.Millisecond * 50)
	//span.LogEvent("halfway done")
	//time.Sleep(time.Millisecond * 50)
}

/***
Tracing
***/
// Initializes an OTLP exporter, and configures the trace provider
func initTracer() func() {
	ctx := context.Background()

	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint("tempo:55680"),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	)
	exp, err := otlp.NewExporter(ctx, driver)
	handleErr(err, "failed to create exporter")

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("demo-service"),
		),
	)
	handleErr(err, "failed to create resource")

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)

	// global vars
	tracer = otel.Tracer("demo-app")
	httpClient = http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	return func() {
		// Shutdown will flush any remaining spans.
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")
	}
}

/***
Server
***/
func instrumentedServer(handler http.HandlerFunc) *http.Server {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	otelHandler := otelhttp.NewHandler(handler, "request")

	tracingMiddleware := func(w http.ResponseWriter, r *http.Request) {
		otelHandler.ServeHTTP(w, r)
	}

	return &http.Server{
		Handler: http.HandlerFunc(tracingMiddleware),
		Addr:    "0.0.0.0:8000",
	}
}

/***
Client
***/
func instrumentedGet(ctx context.Context, url string) (*http.Response, error) {
	// create http request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	return httpClient.Do(req)
}

func handleErr(err error, message string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", err, message))
	}
}
