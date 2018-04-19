package main

import (
	"errors"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
)

// Init validates the transition towards the "init" state
func Init(storeReader store.SegmentReader, l *cs.Link) error {
	return errors.New("error")
}
