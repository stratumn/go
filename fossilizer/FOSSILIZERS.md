# Available fossilizer implementations

We offer several implementations using different kinds of proofs.
You can directly use them as a library or use one of our [docker images](https://hub.docker.com/u/stratumn).

## Bitcoin Fossilizer

This implementation uses the Bitcoin blockchain to store merkle roots.
It collects data in a merkle tree during a configurable `interval`, after which
it computes a merkle root and sends that value to the Bitcoin blockchain.
It produces a proof containing the transaction ID and a merkle path.

Note: you need to provide a [WIF](https://en.bitcoin.it/wiki/Wallet_import_format)
with enough satoshis to send transactions to the blockchain.

## Dummy Fossilizer

This implementation simply creates a timestamp of the request.
It doesn't offer any kind of cryptographic proof and should only be used when
prototyping.

## Dummy Batch Fossilizer

This implementation batches incoming requests in a merkle tree and
asynchronously creates a timestamp for the tree root.
It doesn't offer any kind of cryptographic proof and should only be used when
prototyping with batch systems.
