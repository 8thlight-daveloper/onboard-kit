package encoder_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestFirstOpenedAtIsRepresentationalText(t *testing.T) {
	for _, c := range []struct{ name, text, line string }{
		{"unknown", "", ""},
		{"timestamp", "2026-10-23T12:34:56Z", "firstOpenedAt: \"2026-10-23T12:34:56Z\"\n"},
		{"not-a-time", "not a timestamp", "firstOpenedAt: not a timestamp\n"},
		{"reserved", "deferred", "firstOpenedAt: \"deferred\"\n"},
		{"spaces", " held ", "firstOpenedAt: \" held \"\n"},
		{"line-separators", "first\u2028second\u2029third\nend", "firstOpenedAt: \"first\\u2028second\\u2029third\\nend\"\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			record := model.Record{SchemaVersion: 1, StartDate: "legacy date", FirstOpenedAt: model.Timestamp(c.text), Plan: model.PlanReference{ID: "plan-id"}}
			got, problems := newEncoder().Encode(record)
			want := "schemaVersion: 1\nkind: record\nstartDate: legacy date\n" + c.line + "plan:\n  id: plan-id\n"
			if len(problems) != 0 || string(got) != want {
				t.Fatalf("got %q and %+v, want %q", got, problems, want)
			}
		})
	}
}

func TestFirstOpenedAtCollisionIsRejectedEvenWhenTheFieldIsEmpty(t *testing.T) {
	for _, held := range []model.Timestamp{"", "held event"} {
		record := model.Record{SchemaVersion: 1, FirstOpenedAt: held, UnknownFields: map[string]any{"firstOpenedAt": make(chan int)}}
		got, problems := newEncoder().Encode(record)
		want := []expectedProblem{{Code: "encode.unknown-field-collision", Path: "firstOpenedAt"}}
		if len(got) != 0 || !reflect.DeepEqual(problemsOf(problems), want) || problems[0].Severity != model.SeverityError {
			t.Fatalf("got %q and %+v, want collision only", got, problems)
		}
	}
}

func TestRecordUnicodeSeparatorsUseExplicitEscapes(t *testing.T) {
	record := model.Record{
		SchemaVersion: 1, Account: "a\u2028b", StartDate: "c\u2029d",
		Plan:          model.PlanReference{ID: "e\u2028f", Hash: "g\u2029h", URL: "i\n\u2028j"},
		Answers:       model.Answers{Plan: map[string]model.Value{"q\u2028r": {Kind: model.ValueKindText, Text: "s\u2029t"}, "list": {Kind: model.ValueKindList, List: []string{"u\u2028v"}}}},
		Outcomes:      []model.Outcome{{Section: "w\u2029x", Kind: "y\u2028z", On: "a\u2029b", SectionHash: "c\u2028d", Note: "e\n\u2029f", Differently: "g\u2028h"}},
		UnknownFields: map[string]any{"x\u2028key": map[string]any{"nested\u2029key": []any{"value\u2028text"}}},
	}
	want := "schemaVersion: 1\nkind: record\naccount: \"a\\u2028b\"\nstartDate: \"c\\u2029d\"\nplan:\n  id: \"e\\u2028f\"\n  hash: \"g\\u2029h\"\n  url: \"i\\n\\u2028j\"\nanswers:\n  plan:\n    list:\n      - \"u\\u2028v\"\n    \"q\\u2028r\": \"s\\u2029t\"\noutcomes:\n  - section: \"w\\u2029x\"\n    outcome: \"y\\u2028z\"\n    on: \"a\\u2029b\"\n    sectionHash: \"c\\u2028d\"\n    note: \"e\\n\\u2029f\"\n    differently: \"g\\u2028h\"\n\"x\\u2028key\":\n  \"nested\\u2029key\":\n    - \"value\\u2028text\"\n"
	got, problems := newEncoder().Encode(record)
	if len(problems) != 0 || string(got) != want {
		t.Fatalf("got %q and %+v, want %q", got, problems, want)
	}
}
