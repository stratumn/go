# Available store implementations

We offer several implementations using different kinds of database.
You can directly use them as a library or use one of our [docker images](https://hub.docker.com/u/stratumn).

## Postgres Store

This implementation uses [PostgreSQL](https://www.postgresql.org/) as the underlying database.

This is the most robust implementation we currently offer.
This is what we recommend if you plan on building production-ready applications
that leverage ChainScript.

## File Store

This implementation uses files for storing the data.
You should only use it when doing some prototyping that needs data persistence.

## Dummy Store

This implementation keeps the data in RAM.
You should only use it when doing some prototyping.

## Couch Store

This implementation uses [CouchDB](http://couchdb.apache.org/).

If you're interested in using it, you should probably contribute to help make
it production-ready.

## Rethink Store

This implementation uses [RethinkDB](https://www.rethinkdb.com/).

If you're interested in using it, you should probably contribute to help make
it production-ready.

## ElasticSearch Store

This implementation uses [ElasticSearch](https://www.elastic.co/products/elasticsearch).

If you're interested in using it, you should probably contribute to help make
it production-ready.
