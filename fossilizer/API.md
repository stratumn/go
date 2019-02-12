# HTTP API

The fossilizer exposes a REST endpoint with the following APIs.

## GET /

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

## POST /fossils

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
