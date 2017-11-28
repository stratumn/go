package validator

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"

	log "github.com/sirupsen/logrus"
)

type rootValidator struct {
	hash                *types.Bytes32
	ValidByDefault      bool
	ValidatorsByProcess map[string][]selectiveValidator
}

type selectiveValidator interface {
	validator
	Filter(store.Reader, *cs.Segment) bool
	FilterLink(store.SegmentReader, *cs.Link) bool
}

type jsonSchemaData []struct {
	Type   string           `json:"type"`
	Schema *json.RawMessage `json:"schema"`
}

// NewRootValidator creates a validator from JSON schema filename
func NewRootValidator(filename string, validByDefault bool) Validator {
	v := rootValidator{ValidByDefault: validByDefault}

	log.Debug("loading validator %s", filename)
	f, err := os.Open(filename)
	if err != nil {
		log.Error(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error(err)
	}
	if err = v.loadFromJSON(data); err != nil {
		log.Error(err)
	}

	if data != nil && len(data) > 0 {
		v.updateHash(data)
	}

	return &v
}

func (rv rootValidator) Validate(store store.Reader, segment *cs.Segment) error {
	validByDefault := rv.ValidByDefault
	processValidators, exists := rv.ValidatorsByProcess[segment.Link.GetProcess()]
	if !exists && !validByDefault {
		return errors.New("root validation failed : process validation not found")
	}

	for _, validator := range processValidators {
		if validator.Filter(store, segment) {
			if err := validator.Validate(store, segment); err != nil {
				return err
			}
			validByDefault = true
		}
	}
	if !validByDefault {
		return errors.New("root validation failed")
	}
	return nil
}

func (rv rootValidator) ValidateLink(store store.SegmentReader, link *cs.Link) error {
	validByDefault := rv.ValidByDefault
	processValidators, exists := rv.ValidatorsByProcess[link.GetProcess()]
	if !exists && !validByDefault {
		return errors.New("root validation failed : process validation not found")
	}

	for _, validator := range processValidators {
		if validator.FilterLink(store, link) {
			if err := validator.ValidateLink(store, link); err != nil {
				return err
			}
			validByDefault = true
		}
	}
	if !validByDefault {
		return errors.New("root validation failed")
	}
	return nil
}

func (rv rootValidator) Hash() *types.Bytes32 {
	return rv.hash
}

func (rv *rootValidator) loadFromJSON(data []byte) error {
	var jsonStruct map[string]jsonSchemaData
	err := json.Unmarshal(data, &jsonStruct)

	if err != nil {
		return err
	}

	rv.ValidatorsByProcess = make(map[string][]selectiveValidator, len(jsonStruct))
	for processName, jsonSchemaData := range jsonStruct {
		var actionValidators = make([]selectiveValidator, len(jsonSchemaData))
		for i, val := range jsonSchemaData {
			if val.Schema == nil {
				return fmt.Errorf("loadFromJSON: schema missing for validator %v", val)
			}

			if val.Type == "" {
				return fmt.Errorf("loadFromJSON: type missing for validator %v", val)
			}

			schemaData, _ := val.Schema.MarshalJSON()

			sv, err := newSchemaValidator(val.Type, schemaData)
			if err != nil {
				return err
			}

			actionValidators[i] = sv
		}
		rv.ValidatorsByProcess[processName] = actionValidators
	}

	log.Debugf("validators loaded: %d", len(rv.ValidatorsByProcess))

	return nil
}

func (rv *rootValidator) updateHash(data []byte) {
	byteHash := sha256.Sum256(data)
	validationsHash := types.Bytes32(byteHash)
	rv.hash = &validationsHash
}
