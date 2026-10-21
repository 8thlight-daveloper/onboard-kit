package encoder_test

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

type fixtureCase struct {
	Name     string            `json:"name"`
	Input    string            `json:"input"`
	Output   string            `json:"output"`
	Problems []expectedProblem `json:"problems"`
	Source   string            `json:"source"`
}

func loadCases(t *testing.T) []fixtureCase {
	t.Helper()
	data := readFixture(t, "cases.json")
	var cases []fixtureCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("cases.json: %v", err)
	}
	return cases
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

func loadRecord(t *testing.T, name string) model.Record {
	t.Helper()
	record, err := jsonform.Record(readFixture(t, name))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return record
}

func problemsOf(problems []model.Problem) []expectedProblem {
	var actual []expectedProblem
	for _, p := range problems {
		actual = append(actual, expectedProblem{Code: p.Code, Path: p.Detail["path"]})
	}
	return actual
}

func TestEveryCaseCitesItsSource(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
	}
}

func TestEncodesTheCanonicalFile(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output == "" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newEncoder().Encode(loadRecord(t, c.Input))
			if len(problems) != 0 {
				t.Fatalf("unexpected problems: %+v", problems)
			}
			want := readFixture(t, c.Output)
			if !bytes.Equal(got, want) {
				t.Errorf("bytes differ\n--- got ---\n%s\n--- want ---\n%s", got, want)
			}
		})
	}
}

func TestReportsProblemsAndWritesNothing(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output != "" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newEncoder().Encode(loadRecord(t, c.Input))
			if len(got) != 0 {
				t.Errorf("want no bytes, got %q", got)
			}
			for _, p := range problems {
				if p.Severity != model.SeverityError {
					t.Errorf("problem %s has severity %q, want error", p.Code, p.Severity)
				}
			}
			if actual := problemsOf(problems); !reflect.DeepEqual(actual, c.Problems) {
				t.Errorf("problems differ\n got: %+v\nwant: %+v", actual, c.Problems)
			}
		})
	}
}

func TestWritesTheKindOnEveryRecord(t *testing.T) {
	got, problems := newEncoder().Encode(model.Record{SchemaVersion: 1})
	if len(problems) != 0 || string(got) != "schemaVersion: 1\nkind: record\n" {
		t.Errorf("got %q and %+v", got, problems)
	}
}

func TestWritesTheSchemaVersionAsHeld(t *testing.T) {
	cases := map[string]struct {
		version int
		want    string
	}{
		"one":  {1, "schemaVersion: 1\nkind: record\n"},
		"two":  {2, "schemaVersion: 2\nkind: record\n"},
		"zero": {0, "schemaVersion: 0\nkind: record\n"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, problems := newEncoder().Encode(model.Record{SchemaVersion: c.version})
			if len(problems) != 0 || string(got) != c.want {
				t.Errorf("got %q and %+v, want %q", got, problems, c.want)
			}
		})
	}
}

func TestAnEmptyMapAndANilMapWriteTheSame(t *testing.T) {
	nilMaps := model.Record{SchemaVersion: 1}
	emptyMaps := model.Record{
		SchemaVersion: 1,
		Answers:       model.Answers{Plan: map[string]model.Value{}, Sections: map[string]map[string]model.Value{"s": {}}},
		Outcomes:      []model.Outcome{},
		UnknownFields: map[string]any{},
	}
	first, _ := newEncoder().Encode(nilMaps)
	second, _ := newEncoder().Encode(emptyMaps)
	if !bytes.Equal(first, second) || string(first) != "schemaVersion: 1\nkind: record\n" {
		t.Errorf("got %q and %q", first, second)
	}
}

func TestWritesAnOutcomeWordAsHeldWithoutCheckingIt(t *testing.T) {
	cases := map[string]struct {
		kind model.OutcomeKind
		want string
	}{
		"verified": {model.OutcomeVerified, "    outcome: verified\n"},
		"failed":   {model.OutcomeFailed, "    outcome: failed\n"},
		"deferred": {model.OutcomeDeferred, "    outcome: deferred\n"},
		"unknown":  {"started", "    outcome: started\n"},
		"reserved": {"yes", "    outcome: \"yes\"\n"},
		"empty":    {"", ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			record := model.Record{SchemaVersion: 1, Outcomes: []model.Outcome{{Section: "s", Kind: c.kind}}}
			got, problems := newEncoder().Encode(record)
			want := "schemaVersion: 1\nkind: record\noutcomes:\n  - section: s\n" + c.want
			if len(problems) != 0 || string(got) != want {
				t.Errorf("got %q and %+v, want %q", got, problems, want)
			}
		})
	}
}

func TestWritesAnAnswerOfEveryKindInAnOutcomeFreeRecord(t *testing.T) {
	record := model.Record{SchemaVersion: 1, Answers: model.Answers{Plan: map[string]model.Value{
		"d": {Kind: model.ValueKindDeferred},
		"f": {Kind: model.ValueKindFlag},
		"l": {Kind: model.ValueKindList},
		"t": {Kind: model.ValueKindText, Text: "deferred"},
	}}}
	got, problems := newEncoder().Encode(record)
	want := "schemaVersion: 1\nkind: record\nanswers:\n  plan:\n    d: deferred\n    f: false\n    l: []\n    t: \"deferred\"\n"
	if len(problems) != 0 || string(got) != want {
		t.Errorf("got %q and %+v, want %q", got, problems, want)
	}
}

func TestReportsAnUnsupportedUnknownFieldValue(t *testing.T) {
	record := model.Record{SchemaVersion: 1, UnknownFields: map[string]any{"x-bad": make(chan int)}}
	got, problems := newEncoder().Encode(record)
	if len(got) != 0 {
		t.Errorf("want no bytes, got %q", got)
	}
	if len(problems) != 1 || problems[0].Code != "encode.unsupported-value" || problems[0].Detail["path"] != "x-bad" {
		t.Errorf("want one encode.unsupported-value at x-bad, got %+v", problems)
	}
}

func TestWritesTheSupportedNumberKinds(t *testing.T) {
	cases := map[string]struct {
		value any
		want  string
	}{
		"int":           {int(5), "schemaVersion: 1\nkind: record\nx: 5\n"},
		"int64":         {int64(-7), "schemaVersion: 1\nkind: record\nx: -7\n"},
		"float64":       {1.5, "schemaVersion: 1\nkind: record\nx: 1.5\n"},
		"whole decimal": {float64(2), "schemaVersion: 1\nkind: record\nx: 2.0\n"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, problems := newEncoder().Encode(model.Record{SchemaVersion: 1, UnknownFields: map[string]any{"x": c.value}})
			if len(problems) != 0 || string(got) != c.want {
				t.Errorf("got %q and %+v, want %q", got, problems, c.want)
			}
		})
	}
}

func TestReportsEveryUnsupportedNumberAndKind(t *testing.T) {
	cases := map[string]struct {
		value any
		path  string
	}{
		"int8":    {int8(1), "x"},
		"uint":    {uint(1), "x"},
		"float32": {float32(1.5), "x"},
		"NaN":     {math.NaN(), "x"},
		"+Inf":    {math.Inf(1), "x"},
		"-Inf":    {math.Inf(-1), "x"},
		"strings": {[]string{"a"}, "x"},
		"struct":  {struct{}{}, "x"},
		"in list": {[]any{"ok", int8(1)}, "x[1]"},
		"in map":  {map[string]any{"k": uint(1)}, "x.k"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, problems := newEncoder().Encode(model.Record{SchemaVersion: 1, UnknownFields: map[string]any{"x": c.value}})
			if len(got) != 0 {
				t.Errorf("want no bytes, got %q", got)
			}
			if len(problems) != 1 || problems[0].Code != "encode.unsupported-value" || problems[0].Detail["path"] != c.path {
				t.Errorf("want one encode.unsupported-value at %s, got %+v", c.path, problems)
			}
		})
	}
}

func TestReportsACollidingFieldOnceWithoutExaminingItsValue(t *testing.T) {
	record := model.Record{SchemaVersion: 1, UnknownFields: map[string]any{"kind": make(chan int), "x-bad": int8(1)}}
	got, problems := newEncoder().Encode(record)
	want := []expectedProblem{{"encode.unknown-field-collision", "kind"}, {"encode.unsupported-value", "x-bad"}}
	if len(got) != 0 || !reflect.DeepEqual(problemsOf(problems), want) {
		t.Errorf("got %q and %+v, want %+v", got, problemsOf(problems), want)
	}
}

func TestReportsEveryKindOfProblemInFileOrder(t *testing.T) {
	record := model.Record{
		SchemaVersion: 1,
		Answers:       model.Answers{Plan: map[string]model.Value{"a": {}}, Sections: map[string]map[string]model.Value{"s": {"b": {}}}},
		Outcomes:      []model.Outcome{{}},
		UnknownFields: map[string]any{"answers": "x", "x-bad": uint(1)},
	}
	got, problems := newEncoder().Encode(record)
	want := []expectedProblem{
		{"encode.empty-answer", "answers.plan.a"},
		{"encode.empty-answer", "answers.sections.s.b"},
		{"encode.empty-item", "outcomes[0]"},
		{"encode.unknown-field-collision", "answers"},
		{"encode.unsupported-value", "x-bad"},
	}
	if len(got) != 0 || !reflect.DeepEqual(problemsOf(problems), want) {
		t.Errorf("got %q and %+v, want %+v", got, problemsOf(problems), want)
	}
}

func TestOutputIsDeterministic(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output == "" {
			continue
		}
		record := loadRecord(t, c.Input)
		first, _ := newEncoder().Encode(record)
		for i := 0; i < 5; i++ {
			again, _ := newEncoder().Encode(record)
			if !bytes.Equal(first, again) {
				t.Fatalf("%s: two calls gave different bytes", c.Name)
			}
		}
	}
}

func TestInputIsNotChanged(t *testing.T) {
	for _, c := range loadCases(t) {
		record := loadRecord(t, c.Input)
		snapshot := loadRecord(t, c.Input)
		newEncoder().Encode(record)
		if !reflect.DeepEqual(record, snapshot) {
			t.Errorf("%s: the record changed during Encode", c.Name)
		}
	}
}

func TestOutputIsWellFormed(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output == "" {
			continue
		}
		got, _ := newEncoder().Encode(loadRecord(t, c.Input))
		text := string(got)
		if !strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\n\n") {
			t.Errorf("%s: must end in exactly one newline", c.Name)
		}
		if strings.Contains("\n"+text, "\n---") {
			t.Errorf("%s: must not contain a document marker", c.Name)
		}
		for i, line := range strings.Split(text, "\n") {
			if line != strings.TrimRight(line, " \t") {
				t.Errorf("%s: line %d has trailing whitespace: %q", c.Name, i+1, line)
			}
			if strings.Contains(line, "\t") {
				t.Errorf("%s: line %d has a tab", c.Name, i+1)
			}
		}
	}
}
