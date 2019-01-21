# Fossilizer

A fossilizer takes arbitrary data and provides an externally-verifiable proof
of existence for that data.
It also provides a relative ordering of the events that produced fossilized data.

For example, a naive Bitcoin fossilizer could hash the given data and include
it in a Bitcoin transaction. Since the Bitcoin blockchain is immutable, it
provides a record that the data existed at block N.

See the Go documentation [here](https://godoc.org/github.com/stratumn/go-core/fossilizer#Adapter).

## Available fossilizer implementations

We offer several implementations using different kinds of proofs.
You can directly use them as a library or use one of our [docker images](https://hub.docker.com/u/stratumn).

### Bitcoin Fossilizer

This implementation uses the Bitcoin blockchain to store merkle roots.
It collects data in a merkle tree during a configurable `interval`, after which
it computes a merkle root and sends that value to the Bitcoin blockchain.
It produces a proof containing the transaction ID and a merkle path.

Note: you need to provide a [WIF](https://en.bitcoin.it/wiki/Wallet_import_format)
with enough satoshis to send transactions to the blockchain.

### Dummy Fossilizer

This implementation simply creates a timestamp of the request.
It doesn't offer any kind of cryptographic proof and should only be used when
prototyping.

### Dummy Batch Fossilizer

This implementation batches incoming requests in a merkle tree and
asynchronously creates a timestamp for the tree root.
It doesn't offer any kind of cryptographic proof and should only be used when
prototyping with batch systems.

## HTTP API

The fossilizer exposes a REST endpoint with the following APIs.

### GET /

Returns basic information about the fossilizer instance.

```http
GET /

{
  "adapter": {
    "name": "batchfossilizer",
    "description": "Stratumn Batch Fossilizer",
    "version": "0.3.0",
    "commit": "f993cff818c9667815417194442b6f5f7dcf6f5c"
  }
}
```

### POST /fossils

Fossilize a small amount of data. We recommend fossilizing a hash of your data.
You can provide human-readable metadata that will be tagged with your fossil.

```http
POST /fossils
{
    "data": "dc0e7d35c008705688692f6b6fb4ada79d78966d63e1132654fe9c4d33f1a391",
    "meta": "Santa Claus is dead"
}

HTTP/1.1 200 OK
```

The asynchronous evidence will look like:

```json
{
  "version": "1.0.0",
  "backend": "batchfossilizer",
  "provider": "batchfossilizer",
  "proof": {
    "merklePath": [
      {
        "left": "dc0e7d35c008705688692f6b6fb4ada79d78966d63e1132654fe9c4d33f1a391",
        "parent": "7cafd69a3092ad735bbb8353ee574db035057edae49facbd4805c529c907f3c5",
        "right": "1d278d38cc6e0d3db6b16bb02783d2fc41c451ba07e2fdbf08b9b84675046c08"
      },
      {
        "left": "7cafd69a3092ad735bbb8353ee574db035057edae49facbd4805c529c907f3c5",
        "parent": "b05e54e11904e31a99bab56970d99898cefa1d1e728a860351b45f2319e50dd8",
        "right": "6394ec9008ad3a67afe3a62c3e51d525b030dd1016297abf5c3a93363f63007d"
      }
    ],
    "merkleRoot": "sF5U4RkE4xqZurVpcNmYmM76HR5yioYDUbRfIxnlDdg=",
    "proof": {
      "data": "3A59NcAIcFaIaS9rb7Stp514lm1j4RMmVP6cTTPxo5E=",
      "timestamp": 1548080666,
      "txid": "6a8c7371a9bcfd33037d4b9f65ce0b81bf1a571e5fc856d947c9c3e3b4827cd1"
    },
    "timestamp": 1548080666
  }
}
```
