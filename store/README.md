# Store

The store interface allows efficient storing and querying of ChainScript data.

Once links have been added to the store, other links will depend on them and
reference their hash. Thus, it's important that links cannot be deleted
otherwise it would invalidate a lot of other links. Updates aren't possible
either for the same reason.

This is reflected in the interface exposed, that focuses only on addition and
structured querying over ChainScript data.

See the Golang documentation [here](https://godoc.org/github.com/stratumn/go-core/store).
