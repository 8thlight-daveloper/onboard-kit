package decoder_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

func TestBootstrapDecoding(t *testing.T) {
	t.Parallel()
	var cases []struct {
		Name, Source, Input string
		Want                json.RawMessage
		Problems            []struct{ Code, Path string }
	}
	if err := json.Unmarshal(readFixture(t, "bootstrap.json"), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("empty bootstrap fixture manifest")
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			if c.Source == "" || (len(c.Want) == 0) == (len(c.Problems) == 0) {
				t.Fatal("fixture requires source and exactly one expectation")
			}
			data := []byte(c.Input)
			decoder := newDecoder()
			got, problems := decoder.Decode(data)
			if string(data) != c.Input {
				t.Fatal("input changed")
			}
			if len(c.Problems) == 0 {
				want, err := jsonform.Plan(c.Want)
				if err != nil {
					t.Fatal(err)
				}
				if len(problems) != 0 || !reflect.DeepEqual(got, want) {
					t.Fatalf("got %#v, problems %#v; want %#v", got, problems, want)
				}
			} else {
				if !reflect.DeepEqual(got, model.Plan{}) || len(problems) != len(c.Problems) {
					t.Fatalf("partial result or wrong problem count: %#v, %#v", got, problems)
				}
				for i, p := range problems {
					if p.Code != c.Problems[i].Code || p.Detail["path"] != c.Problems[i].Path || p.Severity != model.SeverityError || strings.TrimSpace(p.Message) == "" {
						t.Fatalf("problem %d: %#v; want %#v", i, p, c.Problems[i])
					}
				}
			}
			again, repeated := decoder.Decode(data)
			if !reflect.DeepEqual(got, again) || !reflect.DeepEqual(problems, repeated) {
				t.Fatal("nondeterministic")
			}
		})
	}
}

func TestBootstrapStateAcrossCalls(t *testing.T) {
	t.Parallel()
	decoder := newDecoder()
	for _, c := range []struct {
		text   string
		marker bool
	}{{"true", true}, {"false", false}, {"null", false}, {"true", true}, {"", false}} {
		data := []byte("schemaVersion: 1\nsections:\n  - id: access\n    bootstrap: " + c.text + "\n")
		got, problems := decoder.Decode(data)
		want := model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "access", Bootstrap: c.marker}}}
		if len(problems) != 0 || !reflect.DeepEqual(got, want) {
			t.Fatalf("marker retained across inputs: %#v, %#v", got, problems)
		}
	}
}
