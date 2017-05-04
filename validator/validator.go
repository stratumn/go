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
	"github.com/xeipuuv/gojsonschema"
)

type Validator interface {
	Validate(store.Batch, *cs.Segment) error
}

type selectiveValidator interface {
	Validator
	Select(store.Batch, *cs.Segment) bool
}

type rootValidator struct {
	ValidateByDefault bool
	Validators        []selectiveValidator
}

type schemaValidator struct {
	Type   string
	Schema []byte
}

type jsonData []struct {
	Type   string           `json:"type"`
	Schema *json.RawMessage `json:"schema"`
}

func (sv rootValidator) Validate(adapter store.Batch, segment *cs.Segment) error {
	validateByDefault := sv.ValidateByDefault
	for _, validator := range sv.Validators {
		if validator.Select(adapter, segment) {
			if err := validator.Validate(adapter, segment); err != nil {
				return err
			}
			validateByDefault = true
		}
	}
	if !validateByDefault {
		return errors.New("root validation failed")
	}
	return nil
}

func (sv schemaValidator) Validate(adapter store.Batch, segment *cs.Segment) error {

	segmentBytes, _ := json.Marshal(segment.Link.State)
	schema := gojsonschema.NewBytesLoader(sv.Schema)
	segmentData := gojsonschema.NewBytesLoader(segmentBytes)

	result, err := gojsonschema.Validate(schema, segmentData)

	if err != nil {
		return err
	}

	if !result.Valid() {
		return errors.New(fmt.Sprintf("segment validation failed: %s", result.Errors()))
	}
	return nil
}

func (sv schemaValidator) Select(adapter store.Batch, segment *cs.Segment) bool {

	// TODO: standartise action as string
	segmentAction, ok := segment.Link.Meta["action"].(string)
	if !ok {
		log.Error("malformed segment")
		return false
	}

	if segmentAction != sv.Type {
		return false
	}

	return true
}

func NewRootValidator(filename string, validateByDefault bool) *rootValidator {
	v := rootValidator{ValidateByDefault: validateByDefault}

	log.Debug("loading validator %s", filename)
	f, err := os.Open(filename)
	if err != nil {
		log.Error(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error(err)
	}
	if err = v.LoadFromJSON(data); err != nil {
		log.Error(err)
	}

	return &v
}

func (rv *rootValidator) LoadFromJSON(data []byte) error {
	var jsonStruct jsonData
	err := json.Unmarshal(data, &jsonStruct)

	if err != nil {
		return err
	}

	rv.Validators = make([]selectiveValidator, len(jsonStruct))
	for i, val := range jsonStruct {
		data, err := val.Schema.MarshalJSON()
		if err != nil {
			return err
		}
		rv.Validators[i] = schemaValidator{Schema: data, Type: val.Type}
	}

	log.Infof("validators loaded: %d", len(rv.Validators))

	return nil

}
