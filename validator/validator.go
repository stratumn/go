package validator

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
)

const (
	// DefaultFilename is the default filename for the file with the rules of validation
	DefaultFilename = "/data/validation/rules.json"
)

// Validator defines the interface with single Validate() method
type Validator interface {
	Validate(store.Reader, *cs.Segment) error
}
