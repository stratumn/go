package validator

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store/storetesting"

	"github.com/xeipuuv/gojsonschema"
)

const defaultJSON = `
[
  {
    "type": "init",
    "schema": {
      "type": "object",
      "properties": {
        "seller": {
          "type": "string"
        },
        "lot": {
          "type": "string"
        },
        "initialPrice": {
          "type": "integer",
          "minimum": 0
        }
      },
      "required": [
        "seller",
        "lot",
        "initialPrice"
      ]
    }
  },
  {
    "type": "bid",
    "schema": {
      "type": "object",
      "properties": {
        "buyer": {
          "type": "string"
        },
        "bidPrice": {
          "type": "integer",
          "minimum": 0
        }
      },
      "required": [
        "buyer",
        "bidPrice"
      ]
    }
  }
]
`

const bidValidator = `
{
  "type": "object",
  "properties": {
    "buyer": {
      "type": "string"
    },
    "bidPrice": {
      "type": "integer",
      "minimum": 0
    }
  },
  "required": [
    "buyer",
    "bidPrice"
  ]
}`

var malformedJSONs = []string{
	``,
	`abba`,
	`{}`,
	`[{}]`,
	`[{"schema":{}}]`,
	`[{"type":"abc"}]`,
	`[{"schema":[], "type":"abc"}]`,
	`[{"schema":"xyz", "type":"abc"}]`,
	`[{"schema":10, "type":"abc"}]`,
	`[{"schema":null, "type":"abc"}]`,
	`[{"schema":{}, "type":{}}]`,
	`[{"schema":{}, "type":10}]`,
	`[{"schema":{}, "type":[]}]`,
	`[{"schema":{}, "type":null}]`,
}

var validJSONs = []struct {
	Data  string
	Count int
}{
	{`[]`, 0},
	{`[{"type":"abc", "schema":{}}]`, 1},
	{`[{"type":"abc", "schema":{}},{"type":"def", "schema":{}}]`, 2},
}

func TestLoadDefaultJSON(t *testing.T) {
	rootValidator := rootValidator{}

	rootValidator.loadFromJSON([]byte(defaultJSON))

	if len(rootValidator.Validators) != 2 {
		t.Errorf("cannot load validators, want 2, got %d", len(rootValidator.Validators))
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	rootValidator := rootValidator{}

	for _, jsonData := range malformedJSONs {
		err := rootValidator.loadFromJSON([]byte(jsonData))

		if err == nil {
			t.Errorf("malformed JSON: error not catched: %s", jsonData)
		}
	}
}

func TestLoadValidJSON(t *testing.T) {
	for _, testCase := range validJSONs {
		rootValidator := rootValidator{}

		err := rootValidator.loadFromJSON([]byte(testCase.Data))

		if err != nil {
			t.Errorf("valid JSON: error: %s, %s", testCase.Data, err)
		}

		if len(rootValidator.Validators) != testCase.Count {
			t.Errorf("valid JSON: validators count mismatch. want: %d, got: %d", testCase.Count, len(rootValidator.Validators))
		}
	}
}

func makeSegment(action string) *cs.Segment {
	return &cs.Segment{Link: cs.Link{
		Meta:  map[string]interface{}{"action": action},
		State: map[string]interface{}{},
	}}
}

func TestFilter(t *testing.T) {
	initSegment := makeSegment("init")
	proposeSegment := makeSegment("propose")

	acceptAllSchema, _ := gojsonschema.NewSchema(gojsonschema.NewBytesLoader([]byte("{}")))

	svInit := schemaValidator{Type: "init", Schema: acceptAllSchema}
	svPropose := schemaValidator{Type: "propose", Schema: acceptAllSchema}

	if !svInit.Filter(&storetesting.MockBatch{}, initSegment) {
		t.Errorf("error not selecting segment `init` by validator of type `init`")
	}

	if svPropose.Filter(&storetesting.MockBatch{}, initSegment) {
		t.Errorf("error selecting segment `init` by validator of type `propose`")
	}

	if svInit.Filter(&storetesting.MockBatch{}, proposeSegment) {
		t.Errorf("error selecting segment `propose` by validator of type `init`")
	}

	initSegment.Link.Meta["action"] = 10

	if svInit.Filter(&storetesting.MockBatch{}, initSegment) {
		t.Errorf("error selecting incorrect segment")
	}
}

func TestSchemaValidate(t *testing.T) {
	bidValidSegment := makeSegment("bid")
	bidInvalidSegment := makeSegment("bid")

	bidValidSegment.Link.State["buyer"] = "Alice"
	bidValidSegment.Link.State["bidPrice"] = 10

	defaultSchema, _ := gojsonschema.NewSchema(gojsonschema.NewBytesLoader([]byte(bidValidator)))

	svBid := schemaValidator{Type: "bid", Schema: defaultSchema}

	if err := svBid.Validate(&storetesting.MockBatch{}, bidValidSegment); err != nil {
		t.Errorf("error not validating valid segment: %s", err)
	}

	if err := svBid.Validate(&storetesting.MockBatch{}, bidInvalidSegment); err == nil {
		t.Errorf("error validating invalid segment `bid`")
	}
}

func TestNewRootValidator(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "go-test")
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(defaultJSON)); err != nil {
		t.Error(err)
	}

	defaultRootValidator := NewRootValidator(tmpfile.Name(), true)

	if len(defaultRootValidator.(*rootValidator).Validators) != 2 {
		t.Errorf("fail to load root validator")
	}

	if err := tmpfile.Close(); err != nil {
		t.Error(err)
	}

	fileNotFoundRootValidator := NewRootValidator("/file/that/does/not/exist", true)

	if len(fileNotFoundRootValidator.(*rootValidator).Validators) != 0 {
		t.Errorf("fail to create root validator: file not found")
	}
}

func TestRootValidator(t *testing.T) {
	segment := makeSegment("init")

	acceptAllSchema, _ := gojsonschema.NewSchema(gojsonschema.NewBytesLoader([]byte("{}")))
	acceptNoneSchema, _ := gojsonschema.NewSchema(gojsonschema.NewBytesLoader([]byte(`{"type": "array"}`)))
	sv1 := schemaValidator{Type: "init", Schema: acceptAllSchema}
	sv2 := schemaValidator{Type: "init", Schema: acceptNoneSchema}

	validators1 := make([]selectiveValidator, 0)
	validators1 = append(validators1, sv1)
	validators2 := make([]selectiveValidator, 0)
	validators2 = append(validators2, sv2)
	rv1 := rootValidator{Validators: validators1, ValidByDefault: true}
	rv2 := rootValidator{Validators: validators2, ValidByDefault: true}
	rv3 := rootValidator{ValidByDefault: false}

	if err := rv1.Validate(&storetesting.MockBatch{}, segment); err != nil {
		t.Errorf("failed to validate rv1")
	}
	if err := rv2.Validate(&storetesting.MockBatch{}, segment); err == nil {
		t.Errorf("rv2 validation successeful")
	}
	if err := rv3.Validate(&storetesting.MockBatch{}, segment); err == nil {
		t.Errorf("rv3 validation successeful")
	}
}
