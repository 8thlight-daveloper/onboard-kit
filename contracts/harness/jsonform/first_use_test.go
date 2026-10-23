package jsonform_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

func TestRecordFirstOpenedAtReaderHasAnIndependentOracle(t *testing.T) {
	for _, c := range []struct {
		input string
		want  model.Timestamp
	}{
		{`{"schemaVersion":1,"startDate":"legacy","firstOpenedAt":" held\nnot a time\u2028\u2029 "}`, " held\nnot a time\u2028\u2029 "},
		{`{"schemaVersion":1,"startDate":"legacy","firstOpenedAt":""}`, ""},
		{`{"schemaVersion":1,"startDate":"legacy"}`, ""},
	} {
		got, err := jsonform.Record([]byte(c.input))
		want := model.Record{SchemaVersion: 1, StartDate: "legacy", FirstOpenedAt: c.want}
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("got %#v and %v, want %#v", got, err, want)
		}
	}
}

func TestRecordFirstOpenedAtWriterHasAnIndependentOracle(t *testing.T) {
	for _, held := range []model.Timestamp{"", " held\nnot a time\u2028\u2029 "} {
		data, err := jsonform.RecordJSON(model.Record{SchemaVersion: 1, StartDate: "legacy", FirstOpenedAt: held})
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{"schemaVersion": float64(1), "startDate": "legacy"}
		if held != "" {
			want["firstOpenedAt"] = string(held)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %s, want %#v", data, want)
		}
	}
}
