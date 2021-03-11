# tempo-otel-example

This example shows a mixture of OpenTelemetry and Prometheus instrumentation in Go.  

It includes trace discovery through Logs and Exemplars.

## Steps

1. Install Loki Docker Driver

`docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions`

2. Build the application.

`./build.sh`

3. Start everything up.

`docker-compose up -d`

4. Exercise the application.

`curl http://localhost:8000/`

5. (Optional) See OpenMetrics Exemplars.

`curl -H 'Accept: application/openmetrics-text' http://localhost:8000/metrics | less`

6. Discover Traces

From Exemplars:

- Navigate to http://localhost:3000/explore
- Go to the explore page
- Choose the Prometheus datasource and execute `histogram_quantile(.99, sum(rate(demo_request_latency_seconds_bucket[1m])) by (le))`

![Exemplars](./exemplar.png)

From Logs:

- Choose the Loki datasource and execute `{container_name="tempootelexample_tracing-example_1"} | logfmt | latency > 1s`

![Loki Derived Fields](./loki.png)