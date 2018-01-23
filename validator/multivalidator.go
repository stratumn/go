package validator

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// MultiValidatorConfig sets the behavior of the validator.
// Its hash can be used to know which validations were applied to a block.
type MultiValidatorConfig struct {
	*multiValidatorConfig
}

// We need an unexported type with exported fields to be able to hash it
// properly without exposing it outside the package.
type multiValidatorConfig struct {
	SchemaConfigs []*schemaValidatorConfig
}

type multiValidator struct {
	config     *MultiValidatorConfig
	validators []validator
}

// NewMultiValidator creates a validator that will simply be a collection
// of single-purpose validators.
func NewMultiValidator(config *MultiValidatorConfig) Validator {
	if config == nil {
		return &multiValidator{}
	}

	var v []validator
	for _, schemaCfg := range config.SchemaConfigs {
		v = append(v, newSchemaValidator(schemaCfg))
	}

	return &multiValidator{
		config:     config,
		validators: v,
	}
}

func (v multiValidator) Hash() *types.Bytes32 {
	b, _ := json.Marshal(v.config)
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash
}

func (v multiValidator) Validate(r store.SegmentReader, l *cs.Link) error {
	for _, child := range v.validators {
		err := child.Validate(r, l)
		if err != nil {
			return err
		}
	}

	return nil
}
