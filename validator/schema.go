package validator

import (
	"encoding/json"
	"fmt"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"

	log "github.com/Sirupsen/logrus"
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

func (sv schemaValidator) Filter(_ store.Reader, segment *cs.Segment) bool {
	// TODO: standardise action as string
	segmentAction, ok := segment.Link.Meta["action"].(string)
	if !ok {
		log.Debug("No action found in segment %v", segment)
		return false
	}

	if segmentAction != sv.Type {
		return false
	}

	return true
}

func (sv schemaValidator) Validate(_ store.Reader, segment *cs.Segment) error {
	segmentBytes, err := json.Marshal(segment.Link.State)
	if err != nil {
		return err
	}

	segmentData := gojsonschema.NewBytesLoader(segmentBytes)

	result, err := sv.Schema.Validate(segmentData)
	if err != nil {
		return err
	}

	if !result.Valid() {
		return fmt.Errorf("segment validation failed: %s", result.Errors())
	}
	return nil
}
