package validator

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store/storetesting"

	"github.com/xeipuuv/gojsonschema"
)

const testProcessName = "testProcess"

var defaultJSON = fmt.Sprintf(`
{
  "%s": [
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
  ],
  "otherProcess": [{
	"type": "abc",
	"schema": {}    
  }]
}
`, testProcessName)

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
	`[]`,
	`[{}]`,
	`{"process": [{"schema":{}}]}`,
	`{"process": [{"type":"abc"}]}`,
	`{"process": [{"schema":[], "type":"abc"}]}`,
	`{"process": [{"schema":"xyz", "type":"abc"}]}`,
	`{"process": [{"schema":10, "type":"abc"}]}`,
	`{"process": [{"schema":null, "type":"abc"}]}`,
	`{"process": [{"schema":{}, "type":{}}]}`,
	`{"process": [{"schema":{}, "type":10}]}`,
	`{"process": [{"schema":{}, "type":[]}]}`,
	`{"process": [{"schema":{}, "type":null}]}`,
}

var validJSONs = []struct {
	Data         string
	ProcessCount int
	ActionCount  int
}{
	{`{}`, 0, 0},
	{`{ "testProcess": [{"type":"abc", "schema":{}}]}`, 1, 1},
	{`{ "testProcess": [{"type":"abc", "schema":{}},{"type":"def", "schema":{}}], "otherProcess": []}`, 2, 2},
}

func TestLoadDefaultJSON(t *testing.T) {
	rootValidator := rootValidator{}

	rootValidator.loadFromJSON([]byte(defaultJSON))

	if len(rootValidator.ValidatorsByProcess) != 2 {
		t.Errorf("cannot load validators, want 2, got %d", len(rootValidator.ValidatorsByProcess))
	}

	if len(rootValidator.ValidatorsByProcess[testProcessName]) != 2 {
		t.Errorf("cannot load validation schema for process %s, want 2, got %d", testProcessName, len(rootValidator.ValidatorsByProcess))
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

		if len(rootValidator.ValidatorsByProcess) != testCase.ProcessCount {
			t.Errorf("valid JSON: validators count mismatch. want: %d, got: %d", testCase.ProcessCount, len(rootValidator.ValidatorsByProcess))
		}

		if len(rootValidator.ValidatorsByProcess[testProcessName]) != testCase.ActionCount {
			t.Errorf("valid JSON: action validators count mismatch. want: %d, got: %d", testCase.ActionCount, len(rootValidator.ValidatorsByProcess[testProcessName]))
		}
	}
}

func makeSegment(action string) *cs.Segment {
	return &cs.Segment{Link: cs.Link{
		Meta:  map[string]interface{}{"action": action, "process": testProcessName},
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

	if len(defaultRootValidator.(*rootValidator).ValidatorsByProcess) != 2 {
		t.Errorf("fail to load root validator")
	}

	if err := tmpfile.Close(); err != nil {
		t.Error(err)
	}

	fileNotFoundRootValidator := NewRootValidator("/file/that/does/not/exist", true)

	if len(fileNotFoundRootValidator.(*rootValidator).ValidatorsByProcess) != 0 {
		t.Errorf("fail to create root validator: file not found")
	}
}

func TestRootValidator(t *testing.T) {
	segment := makeSegment("init")

	acceptAllSchema, _ := gojsonschema.NewSchema(gojsonschema.NewBytesLoader([]byte("{}")))
	acceptNoneSchema, _ := gojsonschema.NewSchema(gojsonschema.NewBytesLoader([]byte(`{"type": "array"}`)))
	sv1 := schemaValidator{Type: "init", Schema: acceptAllSchema}
	sv2 := schemaValidator{Type: "init", Schema: acceptNoneSchema}
	sv3 := schemaValidator{Type: "unknown", Schema: acceptAllSchema}

	validators1 := make(map[string][]selectiveValidator, 0)
	validators1[testProcessName] = make([]selectiveValidator, 0)
	validators1[testProcessName] = append(validators1[testProcessName], sv1)

	validators2 := make(map[string][]selectiveValidator, 0)
	validators2[testProcessName] = make([]selectiveValidator, 0)
	validators2[testProcessName] = append(validators2[testProcessName], sv2)

	validators3 := make(map[string][]selectiveValidator, 0)
	validators3[testProcessName] = make([]selectiveValidator, 0)
	validators3[testProcessName] = append(validators3[testProcessName], sv3)

	rv1 := rootValidator{ValidatorsByProcess: validators1, ValidByDefault: true}
	rv2 := rootValidator{ValidatorsByProcess: validators2, ValidByDefault: true}
	rv3 := rootValidator{ValidByDefault: false}
	rv4 := rootValidator{ValidatorsByProcess: validators3, ValidByDefault: false}

	if err := rv1.Validate(&storetesting.MockBatch{}, segment); err != nil {
		t.Errorf("failed to validate rv1")
	}
	if err := rv2.Validate(&storetesting.MockBatch{}, segment); err == nil {
		t.Errorf("rv2 validation successeful")
	}
	if err := rv3.Validate(&storetesting.MockBatch{}, segment); err == nil {
		t.Errorf("rv3 validation successeful")
	}
	if err := rv4.Validate(&storetesting.MockBatch{}, segment); err == nil {
		t.Errorf("rv4 validation successeful")
	}
}
