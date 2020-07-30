package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaeger_config "github.com/uber/jaeger-client-go/config"
)

var addr = "127.0.0.1:8000"

func main() {
	initJaeger("tracing-example")

	server := instrumentedServer(handler)

	fmt.Println("listening...")
	server.ListenAndServe()
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// make upstream request
	if shouldExecute(80) {
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
	span, _ := opentracing.StartSpanFromContext(ctx, "Long Running Process")
	defer span.Finish()

	time.Sleep(time.Millisecond * 50)
}

/***
Tracing
***/
func initJaeger(service string) {
	// .FromEnv() uses standard environment variables to allow for easy configuration
	//   see docker-compose.yaml
	cfg, err := jaeger_config.FromEnv()
	if err != nil {
		panic(err)
	}

	cfg.InitGlobalTracer(service)
}

/***
Server
***/
func instrumentedServer(handler http.HandlerFunc) *http.Server {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	tracingMiddleware := func(w http.ResponseWriter, r *http.Request) {
		// extract trace context from http
		tracer := opentracing.GlobalTracer()
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

		// create span wrapping handling this http response
		span := tracer.StartSpan("Incoming HTTP Request", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		// inject trace context into Go context
		r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))

		// log traceID
		traceID := span.Context().(jaeger.SpanContext).TraceID()
		logger.Log("traceID", traceID)

		handler(w, r)
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
	// create span wrapping outgoing request
	span, _ := opentracing.StartSpanFromContext(ctx, "Outgoing HTTP Request")
	defer span.Finish()

	// create http request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	// inject trace context into http headers
	span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	return http.DefaultClient.Do(req)
}
