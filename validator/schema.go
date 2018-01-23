package validator

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"

	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

// schemaValidatorConfig contains everything a schemaValidator needs to
// validate links.
type schemaValidatorConfig struct {
	Process string
	Type    string
	Schema  *gojsonschema.Schema
}

// schemaValidator validates the json schema of a link's state.
type schemaValidator struct {
	Config *schemaValidatorConfig
}

// newSchemaValidator creates a schemaValidator for a given process and type.
func newSchemaValidator(process, linkType string, schemaData []byte) (*schemaValidator, error) {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaData))
	if err != nil {
		return nil, err
	}

	config := &schemaValidatorConfig{
		Process: process,
		Type:    linkType,
		Schema:  schema,
	}

	return &schemaValidator{Config: config}, nil
}

// shouldValidate returns true if the link matches the validator's process
// and type. Otherwise the link is considered valid because this validator
// doesn't apply to it.
func (sv schemaValidator) shouldValidate(link *cs.Link) bool {
	linkProcess, ok := link.Meta["process"].(string)
	if !ok {
		log.Debug("No process found in link %v", link)
		return false
	}

	if linkProcess != sv.Config.Process {
		return false
	}

	linkAction, ok := link.Meta["action"].(string)
	if !ok {
		log.Debug("No action found in link %v", link)
		return false
	}

	if linkAction != sv.Config.Type {
		return false
	}

	return true
}

// Validate validates the schema of a link's state.
func (sv schemaValidator) Validate(_ store.SegmentReader, link *cs.Link) error {
	if !sv.shouldValidate(link) {
		return nil
	}

	stateBytes, err := json.Marshal(link.State)
	if err != nil {
		return errors.WithStack(err)
	}

	stateData := gojsonschema.NewBytesLoader(stateBytes)
	result, err := sv.Config.Schema.Validate(stateData)
	if err != nil {
		return errors.WithStack(err)
	}

	if !result.Valid() {
		return fmt.Errorf("link validation failed: %s", result.Errors())
	}

	return nil
}
