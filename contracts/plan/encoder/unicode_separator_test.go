package encoder_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
	"onboard-kit/internal/model"
)

type planUnicodeCase struct {
	name string
	plan func(string) model.Plan
	file string
	path []any
}

func planUnicodeCases() []planUnicodeCase {
	return []planUnicodeCase{
		{"identity", func(s string) model.Plan { return model.Plan{SchemaVersion: 1, ID: s} }, "schemaVersion: 1\nid: @TEXT@\n", []any{"id"}},
		{"provenance", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Provenance: model.Provenance{TemplateSet: s, GeneratedOn: model.Date(s)}}
		}, "schemaVersion: 1\nprovenance:\n  templateSet: @TEXT@\n  generatedOn: @TEXT@\n", []any{"provenance", "generatedOn"}},
		{"origin", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Origin: model.Origin{PlanID: s, Source: s, ContentHash: s, RefreshedOn: model.Date(s)}}
		}, "schemaVersion: 1\norigin:\n  planId: @TEXT@\n  source: @TEXT@\n  contentHash: @TEXT@\n  refreshedOn: @TEXT@\n", []any{"origin", "source"}},
		{"links", func(s string) model.Plan { return model.Plan{SchemaVersion: 1, URL: s, ResultsPath: s} }, "schemaVersion: 1\nplanUrl: @TEXT@\nresultsPath: @TEXT@\n", []any{"resultsPath"}},
		{"contact", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Directory: model.Directory{Contacts: []model.Contact{{ID: s, Name: s, Email: s, Slack: s, SlackID: s}}}}
		}, "schemaVersion: 1\ndirectory:\n  contacts:\n    - id: @TEXT@\n      name: @TEXT@\n      email: @TEXT@\n      slack: @TEXT@\n      slackId: @TEXT@\n", []any{"directory", "contacts", 0, "name"}},
		{"role", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Directory: model.Directory{Roles: []model.Role{{ID: s, Name: s, Holders: []string{s, s}}}}}
		}, "schemaVersion: 1\ndirectory:\n  roles:\n    - id: @TEXT@\n      name: @TEXT@\n      holders:\n        - @TEXT@\n        - @TEXT@\n", []any{"directory", "roles", 0, "holders", 1}},
		{"team-key", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Directory: model.Directory{Teams: []model.Team{{ID: s, Name: s, Contact: s, Holders: map[string][]string{s: {s, "ordinary"}}}}}}
		}, "schemaVersion: 1\ndirectory:\n  teams:\n    - id: @TEXT@\n      name: @TEXT@\n      contact: @TEXT@\n      holders:\n        @TEXT@:\n          - @TEXT@\n          - ordinary\n", []any{"directory", "teams", 0, "holders", "@KEY@", 0}},
		{"variable-strings", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Variables: []model.Variable{{Name: s, Description: s, Example: s, Type: model.VariableType(s), Options: []string{s, "ordinary"}, Pattern: s, PatternHint: s, Default: model.Value{Kind: model.ValueKindText, Text: s}, Audience: model.Audience(s)}}}
		}, "schemaVersion: 1\nvariables:\n  - name: @TEXT@\n    description: @TEXT@\n    example: @TEXT@\n    type: @TEXT@\n    options:\n      - @TEXT@\n      - ordinary\n    pattern: @TEXT@\n    patternHint: @TEXT@\n    default: @TEXT@\n    audience: @TEXT@\n", []any{"variables", 0, "default"}},
		{"variable-default-list", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Variables: []model.Variable{{Name: "selected", Default: model.Value{Kind: model.ValueKindList, List: []string{s, "ordinary", s}}}}}
		}, "schemaVersion: 1\nvariables:\n  - name: selected\n    default:\n      - @TEXT@\n      - ordinary\n      - @TEXT@\n", []any{"variables", 0, "default", 2}},
		{"conditions", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Variables: []model.Variable{{Name: "selected", AskIf: &model.Condition{All: []model.Clause{{Question: s, Answer: model.AnswerTest{Is: model.Value{Kind: model.ValueKindText, Text: s}, IsOneOf: []model.Value{{Kind: model.ValueKindList, List: []string{s, "ordinary"}}}}}, {Section: s, Enabled: false}}}}}}
		}, "schemaVersion: 1\nvariables:\n  - name: selected\n    askIf:\n      all:\n        - question: @TEXT@\n          answer:\n            is: @TEXT@\n            isOneOf:\n              - - @TEXT@\n                - ordinary\n        - section: @TEXT@\n          enabled: false\n", []any{"variables", 0, "askIf", "all", 0, "answer", "is"}},
		{"answer-keys-and-values", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Answers: model.Answers{Plan: map[string]model.Value{s: {Kind: model.ValueKindText, Text: s}}, Sections: map[string]map[string]model.Value{s: {"selection": {Kind: model.ValueKindList, List: []string{s, "ordinary", s}}}}}}
		}, "schemaVersion: 1\nanswers:\n  plan:\n    @TEXT@: @TEXT@\n  sections:\n    @TEXT@:\n      selection:\n        - @TEXT@\n        - ordinary\n        - @TEXT@\n", []any{"answers", "plan", "@KEY@"}},
		{"section-fields", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: s, Title: s, Summary: s, Parent: s, Timing: model.Timing(s), After: []string{s, s}, Requires: []string{s}, Owner: s, Escalation: []string{s}, Verification: s, Body: s, Source: &model.Source{Template: s, Hash: s}}}}
		}, "schemaVersion: 1\nsections:\n  - id: @TEXT@\n    title: @TEXT@\n    summary: @TEXT@\n    parent: @TEXT@\n    timing: @TEXT@\n    after:\n      - @TEXT@\n      - @TEXT@\n    requires:\n      - @TEXT@\n    owner: @TEXT@\n    escalation:\n      - @TEXT@\n    verification: @TEXT@\n    body: @TEXT@\n    source:\n      template: @TEXT@\n      hash: @TEXT@\n", []any{"sections", 0, "body"}},
		{"section-variable", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "activity", Variables: []model.Variable{{Name: s, Description: s, Default: model.Value{Kind: model.ValueKindText, Text: s}}}}}}
		}, "schemaVersion: 1\nsections:\n  - id: activity\n    variables:\n      - name: @TEXT@\n        description: @TEXT@\n        default: @TEXT@\n", []any{"sections", 0, "variables", 0, "name"}},
		{"step-commands-and-map", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "activity", Steps: []model.Step{{ID: s, Name: s, Notes: s, Owner: s, Timing: model.Timing(s), Escalation: []string{s}, Run: []string{s, "ordinary", s}, Open: s, Directory: s, Env: map[string]string{s: s}}}}}}
		}, "schemaVersion: 1\nsections:\n  - id: activity\n    steps:\n      - id: @TEXT@\n        name: @TEXT@\n        notes: @TEXT@\n        owner: @TEXT@\n        timing: @TEXT@\n        escalation:\n          - @TEXT@\n        run:\n          - @TEXT@\n          - ordinary\n          - @TEXT@\n        open: @TEXT@\n        directory: @TEXT@\n        env:\n          @TEXT@: @TEXT@\n", []any{"sections", 0, "steps", 0, "env", "@KEY@"}},
		{"checks", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "activity", Steps: []model.Step{{ID: "verify", Checks: []model.Check{{Run: []string{s, s}}, {Ask: s}, {Env: s}}}}}}}
		}, "schemaVersion: 1\nsections:\n  - id: activity\n    steps:\n      - id: verify\n        checks:\n          - run:\n              - @TEXT@\n              - @TEXT@\n          - ask: @TEXT@\n          - env: @TEXT@\n", []any{"sections", 0, "steps", 0, "checks", 1, "ask"}},
		{"unknown-root-nested-key", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, UnknownFields: map[string]any{s: map[string]any{s: []any{s, "ordinary", s}}}}
		}, "schemaVersion: 1\n@TEXT@:\n  @TEXT@:\n    - @TEXT@\n    - ordinary\n    - @TEXT@\n", []any{"@KEY@", "@KEY@", 2}},
		{"unknown-section-key", func(s string) model.Plan {
			return model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "activity", UnknownFields: map[string]any{s: s}}}}
		}, "schemaVersion: 1\nsections:\n  - id: activity\n    @TEXT@: @TEXT@\n", []any{"sections", 0, "@KEY@"}},
	}
}

func planUnicodeAt(t *testing.T, value any, path []any, text string) any {
	t.Helper()
	for _, part := range path {
		switch key := part.(type) {
		case string:
			if key == "@KEY@" {
				key = text
			}
			object, ok := value.(map[string]any)
			if !ok {
				t.Fatal("literal expected mapping shape differs")
			}
			value = object[key]
		case int:
			items, ok := value.([]any)
			if !ok || key >= len(items) {
				t.Fatal("literal expected list shape differs")
			}
			value = items[key]
		default:
			t.Fatal("invalid independent fixture path")
		}
	}
	return value
}

func TestPlanUnicodeSeparatorsHaveExactCanonicalBytesAndDecodedText(t *testing.T) {
	for _, payload := range []struct{ name, text, quoted string }{
		{"line-separator", "alpha\u2028beta", `"alpha\u2028beta"`},
		{"paragraph-separator", "alpha\u2029beta", `"alpha\u2029beta"`},
		{"already-needs-quotes", " alpha\u2028beta\u2029 ", `" alpha\u2028beta\u2029 "`},
		{"lf-with-both-separators", "alpha\u2028beta\ncharlie\u2029delta\n", `"alpha\u2028beta\ncharlie\u2029delta\n"`},
		{"leading-lf", "\nalpha\u2029beta", `"\nalpha\u2029beta"`},
	} {
		for _, c := range planUnicodeCases() {
			t.Run(payload.name+"/"+c.name, func(t *testing.T) {
				want := []byte(strings.ReplaceAll(c.file, "@TEXT@", payload.quoted))
				var expected map[string]any
				if err := yaml.Unmarshal(want, &expected); err != nil {
					t.Fatalf("independent literal fixture is invalid YAML: %v", err)
				}
				if value := planUnicodeAt(t, expected, c.path, payload.text); value != payload.text {
					t.Fatal("independent literal does not preserve exact original separator text")
				}
				input, snapshot := c.plan(payload.text), c.plan(payload.text)
				encoder := newEncoder()
				got, problems := encoder.Encode(input)
				if len(problems) != 0 {
					t.Errorf("representable separator text refused: %+v", problems)
				}
				if !bytes.Equal(got, want) {
					t.Errorf("exact independently authored canonical bytes differ\n got: %q\nwant: %q", got, want)
				}
				if bytes.ContainsAny(got, "\u2028\u2029") {
					t.Error("Unicode separator emitted literally")
				}
				var actual map[string]any
				if err := yaml.Unmarshal(got, &actual); err != nil {
					t.Errorf("actual separator output cannot be decoded as generic YAML: %v", err)
				} else {
					if !reflect.DeepEqual(actual, expected) {
						t.Error("generic decoded output differs from independent literal mapping")
					}
					if value := planUnicodeAt(t, actual, c.path, payload.text); value != payload.text {
						t.Error("exact original text lost in decoded output")
					}
				}
				if !reflect.DeepEqual(input, snapshot) {
					t.Error("encoder changed held separator text or data")
				}
				again, againProblems := encoder.Encode(c.plan(payload.text))
				if len(againProblems) != 0 || !bytes.Equal(again, want) {
					t.Error("reused encoder changed exact canonical separator output")
				}
			})
		}
	}
}

func TestPlanUnicodeKeyOrderUsesHeldBytesNotEscapedSpelling(t *testing.T) {
	input := model.Plan{SchemaVersion: 1, UnknownFields: map[string]any{"a\u2029b": "second", "z": "last-ascii", "a\u2028b": "first", "é": "ordinary-unicode"}}
	want := []byte("schemaVersion: 1\n\"a\\u2028b\": first\n\"a\\u2029b\": second\nz: last-ascii\n\"é\": ordinary-unicode\n")
	got, problems := newEncoder().Encode(input)
	if len(problems) != 0 || !bytes.Equal(got, want) {
		t.Errorf("exact held-byte map key order differs: %q %+v", got, problems)
	}
	var actual map[string]any
	if err := yaml.Unmarshal(got, &actual); err != nil {
		t.Fatal(err)
	}
	wantMap := map[string]any{"schemaVersion": 1, "a\u2028b": "first", "a\u2029b": "second", "z": "last-ascii", "é": "ordinary-unicode"}
	if !reflect.DeepEqual(actual, wantMap) {
		t.Error("key text changed during escaping")
	}
}

func TestPlanOtherPrintableUnicodeAndOrdinaryLFKeepPriorCanonicalRules(t *testing.T) {
	input := model.Plan{SchemaVersion: 1, ID: "café", Sections: []model.Section{{ID: "activity", Title: "日本語", Body: "first\n日本語\n", Steps: []model.Step{{Run: []string{"echo", "naïve"}}}}}}
	want := []byte("schemaVersion: 1\nid: café\nsections:\n  - id: activity\n    title: \"日本語\"\n    body: |\n      first\n      日本語\n    steps:\n      - run:\n          - echo\n          - naïve\n")
	got, problems := newEncoder().Encode(input)
	if len(problems) != 0 || !bytes.Equal(got, want) {
		t.Errorf("unrelated printable Unicode or ordinary LF canonical form changed: %q %+v", got, problems)
	}
	var actual map[string]any
	if err := yaml.Unmarshal(got, &actual); err != nil {
		t.Fatal(err)
	}
	if actual["id"] != "café" || planUnicodeAt(t, actual, []any{"sections", 0, "body"}, "") != "first\n日本語\n" {
		t.Error("unrelated literal Unicode text changed")
	}
}
