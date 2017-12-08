package validator

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

const (
	// DefaultFilename is the default filename for the file with the rules of validation
	DefaultFilename = "/data/validation/rules.json"
)

// validator defines the interface with single Validate() method
type validator interface {
	Validate(store.SegmentReader, *cs.Link) error
}

// Validator defines a validator that can be identified by a hash
type Validator interface {
	validator
	Hash() *types.Bytes32
}
