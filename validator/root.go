package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"

	log "github.com/Sirupsen/logrus"
)

type rootValidator struct {
	ValidByDefault bool
	Validators     []selectiveValidator
}

type selectiveValidator interface {
	Validator
	Filter(store.Reader, *cs.Segment) bool
}

type jsonData []struct {
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

	return &v
}

func (rv rootValidator) Validate(store store.Reader, segment *cs.Segment) error {
	validByDefault := rv.ValidByDefault
	for _, validator := range rv.Validators {
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

func (rv *rootValidator) loadFromJSON(data []byte) error {
	var jsonStruct jsonData
	err := json.Unmarshal(data, &jsonStruct)

	if err != nil {
		return err
	}

	rv.Validators = make([]selectiveValidator, len(jsonStruct))
	for i, val := range jsonStruct {
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

		rv.Validators[i] = sv
	}

	log.Debugf("validators loaded: %d", len(rv.Validators))

	return nil
}
