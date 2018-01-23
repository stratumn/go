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
	// Validate runs validations on a link and returns an error
	// if the link is invalid.
	Validate(store.SegmentReader, *cs.Link) error
}

// Validator defines a validator that has an internal state, identified by
// its hash.
type Validator interface {
	validator

	// Hash returns the hash of the validator's state.
	// It can be used to know which set of validations were applied
	// to a block.
	Hash() *types.Bytes32
}
