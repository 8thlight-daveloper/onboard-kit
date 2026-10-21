package jsonform_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
)

func TestJSONReadsBackAsTheSamePlan(t *testing.T) {
	var paths []string
	for _, pattern := range []string{"../../plan/testdata/*.plan.json", "../../plan/*/testdata/*.plan.json"} {
		found, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, found...)
	}
	if len(paths) == 0 {
		t.Fatal("no fixtures found")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			plan, err := jsonform.Plan(data)
			if err != nil {
				t.Fatal(err)
			}
			written, err := jsonform.JSON(plan)
			if err != nil {
				t.Fatal(err)
			}
			again, err := jsonform.Plan(written)
			if err != nil {
				t.Fatalf("written JSON does not read back: %v\n%s", err, written)
			}
			if !reflect.DeepEqual(plan, again) {
				t.Errorf("plan changed on the way through JSON\n%s", written)
			}
		})
	}
}

func TestPartsReadAndWriteBack(t *testing.T) {
	section, err := jsonform.Section([]byte(`{"id": "s", "title": "T", "includeIf": {"all": [{"question": "plan.q", "answer": {"isOneOf": ["a", true, {"deferred": true}]}}]}, "after": ["x", "y"]}`))
	if err != nil {
		t.Fatal(err)
	}
	written, err := jsonform.SectionJSON(section)
	if err != nil {
		t.Fatal(err)
	}
	again, err := jsonform.Section(written)
	if err != nil || !reflect.DeepEqual(section, again) {
		t.Errorf("section changed on the way through JSON: %v\n%s", err, written)
	}
	condition, err := jsonform.Condition([]byte(`{"all": [{"question": "plan.q", "answer": {"isSettled": false}}]}`))
	if err != nil || len(condition.All) != 1 || condition.All[0].Answer.IsSettled == nil {
		t.Errorf("condition not read: %v %+v", err, condition)
	}
	value, err := jsonform.Value([]byte(`["a", "b"]`))
	if err != nil || len(value.List) != 2 {
		t.Errorf("value not read: %v %+v", err, value)
	}
	answers, err := jsonform.Answers([]byte(`{"plan": {"q": true}, "sections": {"s": {"n": "x"}}}`))
	if err != nil || len(answers.Plan) != 1 || len(answers.Sections["s"]) != 1 {
		t.Errorf("answers not read: %v %+v", err, answers)
	}
}

func TestRecordJSONReadsBackAsTheSameRecord(t *testing.T) {
	paths, err := filepath.Glob("../../record/testdata/*.record.json")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no record fixtures found: %v", err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			record, err := jsonform.Record(data)
			if err != nil {
				t.Fatal(err)
			}
			written, err := jsonform.RecordJSON(record)
			if err != nil {
				t.Fatal(err)
			}
			again, err := jsonform.Record(written)
			if err != nil {
				t.Fatalf("written JSON does not read back: %v\n%s", err, written)
			}
			if !reflect.DeepEqual(record, again) {
				t.Fatalf("record changed on the way through\nfirst: %#v\nagain: %#v", record, again)
			}
		})
	}
}
