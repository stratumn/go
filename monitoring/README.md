# Configuring Monitoring

Our packages export metrics, logs and traces depending on your configuration.
To activate monitoring, set the `monitoring.active` flag to `true`.

## Logs

Logs are written to the standard output using [logrus](https://github.com/sirupsen/logrus).

## Metrics

We support two ways of exposing metrics:

- [Prometheus](https://prometheus.io/) metrics that can be pulled over tcp
- [Elastic APM](https://www.elastic.co/solutions/apm) metrics that are pushed regularly

### Prometheus

If you want to pull Prometheus metrics, set `monitoring.exporter` to `prometheus`.
When using Prometheus, you should also set `monitoring.metrics.port`.
Your metrics can then be pulled at `{baseURL}:{port}/metrics`.

### Elastic APM

If you want to use Elastic APM, set `monitoring.exporter` to `elastic`.

All the Elastic APM configuration options should be passed as environment
variables.
Please see the [Elastic APM documentation](https://www.elastic.co/guide/en/apm/agent/go/current/configuration.html)
for reference.

The following configuration options are mandatory:

- `ELASTIC_APM_SERVICE_NAME`
- `ELASTIC_APM_SERVER_URL`
- `ELASTIC_APM_SECRET_TOKEN`

We recommend setting the following configuration options too:

- `ELASTIC_APM_ENVIRONMENT` (staging/production)
- `ELASTIC_APM_METRICS_INTERVAL` if the default 30 seconds doesn't suit you

## Traces

Distributed traces are only available when choosing the `elastic` exporter.

If you use distributed tracing, don't forget to set `ELASTIC_APM_TRANSACTION_SAMPLE_RATE`
to less than `1.0` in production.
Distributed tracing is expensive, if you have a lot of requests you can't
afford to trace them all.
