# HTTP API

The store exposes a REST endpoint with the following APIs.

## GET /

Returns basic information about the store instance.

```http
GET /

{
  "adapter": {
    "name": "dummystore",
    "description": "Stratumn's Dummy Store",
    "version": "0.3.0",
    "commit": "f993cff818c9667815417194442b6f5f7dcf6f5c"
  }
}
```

## POST /links

Add a JSON-encoded link to the store.

```http
POST /links
{
    "data": "ewogICJvd25lciI6ICJhbGljZSIKfQ==",
    "meta": {
        "action": "init",
        "clientId": "github.com/stratumn/go-chainscript",
        "mapId": "123456",
        "outDegree": 3,
        "process": { "name": "asset-tracker", "state": "asset-created" },
        "step": "init",
        "tags": ["alice"]
    },
    "version": "1.0.0"
}

HTTP/1.1 200 OK
{
  "link": {
    "version": "1.0.0",
    "data": "ewogICJvd25lciI6ICJhbGljZSIKfQ==",
    "meta": {
      "clientId": "github.com/stratumn/go-chainscript",
      "outDegree": 3,
      "process": { "name": "asset-tracker", "state": "asset-created" },
      "mapId": "123456",
      "action": "init",
      "step": "init",
      "tags": ["alice"]
    }
  },
  "meta": { "linkHash": "z+w01ZMHQ4dyuA1ro5BcKM5NPV6vpgLmZ0XjDTwf7Hw=" }
}
```

## POST /evidences/:linkHash

Add a JSON-encoded evidence to a link (identified by its hex-encoded hash).

```http
POST /evidences/cfec34d59307438772b80d6ba3905c28ce4d3d5eafa602e66745e30d3c1fec7c
{
    "version": "1.0.0",
    "backend": "btc",
    "provider": "testnet:3",
    "proof": "ewogICJ0eGlkIjogImNmZWMzNGQ1OTMwNzQzODc3MmI4MGQ2YmEzOTA1YzI4Y2U0ZDNkNWVhZmE2MDJlNjY3NDVlMzBkM2MxZmVjN2MiCn0="
}

HTTP/1.1 200 OK
{}
```

## GET /segments/:linkHash

Get a link by its hash (hex-encoded).

```http
GET /segments/cfec34d59307438772b80d6ba3905c28ce4d3d5eafa602e66745e30d3c1fec7c

HTTP/1.1 200 OK
{
  "link": {
    "version": "1.0.0",
    "data": "ewogICJvd25lciI6ICJhbGljZSIKfQ==",
    "meta": {
      "clientId": "github.com/stratumn/go-chainscript",
      "outDegree": 3,
      "process": { "name": "asset-tracker", "state": "asset-created" },
      "mapId": "123456",
      "action": "init",
      "step": "init",
      "tags": ["alice"]
    }
  },
  "meta": {
    "linkHash": "z+w01ZMHQ4dyuA1ro5BcKM5NPV6vpgLmZ0XjDTwf7Hw=",
    "evidences": [
      {
        "version": "1.0.0",
        "backend": "btc",
        "provider": "testnet:3",
        "proof": "ewogICJ0eGlkIjogImNmZWMzNGQ1OTMwNzQzODc3MmI4MGQ2YmEzOTA1YzI4Y2U0ZDNkNWVhZmE2MDJlNjY3NDVlMzBkM2MxZmVjN2MiCn0="
      }
    ]
  }
}
```

## GET /segments?[offset=offset]&[limit=limit]&[mapIds[]=id1]&[mapIds[]=id2]&[prevLinkHash=prevLinkHash]&[tags[]=tag1]&[tags[]=tag2]

Search segments using various query string filters.

```http
GET /segments?offset=1&limit=2&tags[]=alice&tags[]=bob

HTTP/1.1 200 OK
{
  "segments": [
    {
      "link": {
        "version": "1.0.0",
        "data": "ewogICJvd25lciI6ICJhbGljZSIKfQ==",
        "meta": {
          "clientId": "github.com/stratumn/go-chainscript",
          "outDegree": 3,
          "process": { "name": "asset-tracker", "state": "asset-created" },
          "mapId": "123456",
          "action": "init",
          "step": "init",
          "tags": ["alice"]
        }
      },
      "meta": {
        "linkHash": "z+w01ZMHQ4dyuA1ro5BcKM5NPV6vpgLmZ0XjDTwf7Hw=",
        "evidences": [
          {
            "version": "1.0.0",
            "backend": "btc",
            "provider": "testnet:3",
            "proof": "ewogICJ0eGlkIjogImNmZWMzNGQ1OTMwNzQzODc3MmI4MGQ2YmEzOTA1YzI4Y2U0ZDNkNWVhZmE2MDJlNjY3NDVlMzBkM2MxZmVjN2MiCn0="
          }
        ]
      }
    },
    {
      "link": {
        "version": "1.0.0",
        "meta": {
          "clientId": "github.com/stratumn/go-chainscript",
          "outDegree": -1,
          "process": { "name": "voting-protocol" },
          "mapId": "234567",
          "action": "create",
          "step": "create",
          "tags": ["bob"]
        }
      },
      "meta": { "linkHash": "9s1Zv+dWfcUzrqUgsbfgeVykRz0tq5bCaPLZMYMOQ4c=" }
    }
  ],
  "totalCount": 5
}
```

## GET /maps?[offset=offset]&[limit=limit]

List map IDs (instance of a process).

```http
GET /maps?limit=25
["123456","234567"]
```

## GET /websocket

Connect to a websocket to receive store events.
The store will send events when links and evidences are added.
