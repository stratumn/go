package main

import (
	"context"
	"errors"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
)

// Init validates the transition towards the "init" state
func Init(storeReader store.SegmentReader, l *cs.Link) error {
	return nil
}

// FetchLink fetches a link and returns a nil error
func FetchLink(storeReader store.SegmentReader, l *cs.Link) error {
	_, err := storeReader.FindSegments(context.Background(), &store.SegmentFilter{
		MapIDs: []string{l.Meta.MapID},
	})
	return err
}

// Invalid validates the transition towards the "invalid" state
func Invalid(storeReader store.SegmentReader, l *cs.Link) error {
	return errors.New("error")
}

// BadSignature is an example of validator which is not of type ScriptValidatorFunc
func BadSignature() error {
	return nil
}

func main() {}
