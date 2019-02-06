# Elastic Stack

We're using [Elastic Stack](https://www.elastic.co/products) to monitor our production services.

This folder contains sample configuration files that lets you try it locally
with Docker. Beware that this is not a production Elastic Stack deployment setup.

## Running the Elastic Stack

Simply run `docker-compose up` to start the Elastic Stack locally.

## Running services

You can then run your services (store, fossilizer) with the following flags set:

- `--monitoring.active=true`
- `--monitoring.exporter=elastic`
- `ELASTIC_APM_SERVICE_NAME=something-that-describes-the-service`
- `ELASTIC_APM_SERVER_URL=http://localhost:8200`
- `ELASTIC_APM_ENVIRONMENT=local`
- `ELASTIC_APM_METRICS_INTERVAL=10s`
- `ELASTIC_APM_TRANSACTION_SAMPLE_RATE=1.0`

For example:

```bash
export ELASTIC_APM_SERVICE_NAME=dummystore
export ELASTIC_APM_SERVER_URL=http://localhost:8200
export ELASTIC_APM_ENVIRONMENT=local
export ELASTIC_APM_METRICS_INTERVAL=10s
export ELASTIC_APM_TRANSACTION_SAMPLE_RATE=1.0

dummystore --monitoring.active=true --monitoring.exporter=elastic
```

## Viewing metrics and traces

Navigate to `http://localhost:5601` to open the Kibana console.

On the front page, click `APM` and click `Load Kibana objects` and then `Launch APM`.
This will enable the APM dashboards.

You can now start making requests to your services and will be able to monitor
everything from Kibana.

## Troubleshooting

### APM Server file permissions

If APM Server fails to start with the following error:

```bash
Exiting: error loading config file: config file ("apm-server.yml") must be owned by the beat user (uid=0) or root
```

You need to set the file permissions on the `/config/apm/apm-server.yml` file:

```bash
sudo chown root ./config/apm/apm-server.yml
```
