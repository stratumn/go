# Fossilizer

A fossilizer takes arbitrary data and provides an externally-verifiable proof
of existence for that data.
It also provides a relative ordering of the events that produced fossilized data.

For example, a naive Bitcoin fossilizer could hash the given data and include
it in a Bitcoin transaction. Since the Bitcoin blockchain is immutable, it
provides a record that the data existed at block N.

See the Go documentation [here](https://godoc.org/github.com/stratumn/go-core/fossilizer#Adapter).
