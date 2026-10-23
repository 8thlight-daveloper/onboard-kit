package decoder_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestFirstOpenedAtDecodesAsExactText(t *testing.T) {
	for _, c := range []struct{ name, value, want string }{
		{"absent", "", ""},
		{"empty", "firstOpenedAt: \"\"\n", ""},
		{"null", "firstOpenedAt: null\n", ""},
		{"blank", "firstOpenedAt:\n", ""},
		{"timestamp", "firstOpenedAt: 2026-10-23T12:34:56Z\n", "2026-10-23T12:34:56Z"},
		{"date", "firstOpenedAt: 2026-10-23\n", "2026-10-23"},
		{"not-a-time", "firstOpenedAt: not a timestamp\n", "not a timestamp"},
		{"reserved-text", "firstOpenedAt: deferred\n", "deferred"},
		{"quoted-number", "firstOpenedAt: \"42\"\n", "42"},
		{"spaces", "firstOpenedAt: ' held '\n", " held "},
		{"separators", "firstOpenedAt: \"a\\u2028b\\u2029c\\nend\"\n", "a\u2028b\u2029c\nend"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, problems := newDecoder().Decode([]byte("schemaVersion: 1\nkind: record\nstartDate: legacy date\n" + c.value))
			want := model.Record{SchemaVersion: 1, StartDate: "legacy date", FirstOpenedAt: model.Timestamp(c.want)}
			if len(problems) != 0 || !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v and %+v, want %#v", got, problems, want)
			}
		})
	}
}

func TestFirstOpenedAtUsesTheTextFormatGuards(t *testing.T) {
	for _, c := range []struct{ name, body, code string }{
		{"boolean", "true", "format.wrong-type"},
		{"list", "[]", "format.wrong-type"},
		{"map", "{}", "format.wrong-type"},
		{"number", "42", "format.number-spelling"},
		{"old-boolean", "yes", "format.boolean-spelling"},
		{"old-null", "~", "format.null-spelling"},
		{"anchor", "&event held", "format.alias"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, problems := newDecoder().Decode([]byte("schemaVersion: 1\nkind: record\nfirstOpenedAt: " + c.body + "\n"))
			want := []expectedProblem{{Code: c.code, Path: "firstOpenedAt"}}
			if !reflect.DeepEqual(got, model.Record{}) || !reflect.DeepEqual(actualProblems(problems), want) {
				t.Fatalf("got %#v and %+v, want zero record and %+v", got, actualProblems(problems), want)
			}
		})
	}
}

func TestFirstOpenedAtDuplicatesAndSourceOrder(t *testing.T) {
	for _, c := range []struct {
		name, body string
		want       []expectedProblem
	}{
		{"duplicate", "firstOpenedAt: one\nfirstOpenedAt: two\n", []expectedProblem{{Code: "format.duplicate-key", Path: "firstOpenedAt"}}},
		{"escaped-key-duplicate", "firstOpenedAt: one\n\"firstOpened\\u0041t\": two\n", []expectedProblem{{Code: "format.duplicate-key", Path: "firstOpenedAt"}}},
		{"source-order", "firstOpenedAt: true\naccount: 42\nstartDate: yes\n", []expectedProblem{{Code: "format.wrong-type", Path: "firstOpenedAt"}, {Code: "format.number-spelling", Path: "account"}, {Code: "format.boolean-spelling", Path: "startDate"}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, problems := newDecoder().Decode([]byte("schemaVersion: 1\nkind: record\n" + c.body))
			if !reflect.DeepEqual(got, model.Record{}) || !reflect.DeepEqual(actualProblems(problems), c.want) {
				t.Fatalf("got %#v and %+v, want zero record and %+v", got, actualProblems(problems), c.want)
			}
		})
	}
}

func TestFirstOpenedAtCaseVariantRemainsUnknown(t *testing.T) {
	got, problems := newDecoder().Decode([]byte("schemaVersion: 1\nkind: record\nfirstOpenedAt: held\nFirstOpenedAt: variant\n"))
	want := model.Record{SchemaVersion: 1, FirstOpenedAt: "held", UnknownFields: map[string]any{"FirstOpenedAt": "variant"}}
	if len(problems) != 0 || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v and %+v, want %#v", got, problems, want)
	}
}

func TestFirstOpenedAtDoesNotBypassVersionOrKindGuards(t *testing.T) {
	for _, c := range []struct {
		name, file string
		want       []expectedProblem
	}{
		{"version", "schemaVersion: 2\nkind: plan\nfirstOpenedAt: true\n", []expectedProblem{{Code: "version.too-new", Path: "schemaVersion", Detail: map[string]string{"found": "2", "supported": "1"}}}},
		{"kind", "schemaVersion: 1\nkind: plan\nfirstOpenedAt: true\n", []expectedProblem{{Code: "kind.not-a-record", Path: "kind", Detail: map[string]string{"found": "plan"}}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, problems := newDecoder().Decode([]byte(c.file))
			if !reflect.DeepEqual(got, model.Record{}) || !reflect.DeepEqual(actualProblems(problems), c.want) {
				t.Fatalf("got %#v and %+v, want zero record and %+v", got, actualProblems(problems), c.want)
			}
		})
	}
}
