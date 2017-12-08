package validator

import (
	"encoding/json"
	"fmt"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"

	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

type schemaValidator struct {
	Type   string
	Schema *gojsonschema.Schema
}

func newSchemaValidator(segmentType string, data []byte) (*schemaValidator, error) {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(data))

	if err != nil {
		return nil, err
	}

	return &schemaValidator{Type: segmentType, Schema: schema}, nil
}

func (sv schemaValidator) Filter(_ store.SegmentReader, link *cs.Link) bool {
	// TODO: standardise action as string
	linkAction, ok := link.Meta["action"].(string)
	if !ok {
		log.Debug("No action found in link %v", link)
		return false
	}

	if linkAction != sv.Type {
		return false
	}

	return true
}

func (sv schemaValidator) Validate(_ store.SegmentReader, link *cs.Link) error {
	stateBytes, err := json.Marshal(link.State)
	if err != nil {
		return err
	}

	stateData := gojsonschema.NewBytesLoader(stateBytes)

	result, err := sv.Schema.Validate(stateData)
	if err != nil {
		return err
	}

	if !result.Valid() {
		return fmt.Errorf("link validation failed: %s", result.Errors())
	}
	return nil
}
