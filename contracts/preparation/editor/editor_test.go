package editor_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/preparation/editor"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

type fixture struct {
	Name        string                     `json:"name"`
	Source      string                     `json:"source"`
	Context     *contract.Context          `json:"context"`
	Current     *model.Preparation         `json:"current"`
	Operation   string                     `json:"operation"`
	Recipient   model.PreparationRecipient `json:"recipient"`
	Arrangement model.Arrangement          `json:"arrangement"`
	ID          string                     `json:"id"`
	Expected    *model.Preparation         `json:"expected"`
	Unchanged   bool                       `json:"unchanged"`
	Problems    []expectedProblem          `json:"problems"`
}

type manifest struct {
	Context contract.Context  `json:"context"`
	Current model.Preparation `json:"current"`
	Cases   []fixture         `json:"cases"`
}

func readManifest(t *testing.T) manifest {
	t.Helper()
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Cases) == 0 {
		t.Fatal("empty fixture manifest")
	}
	seen := make(map[string]bool, len(m.Cases))
	for _, f := range m.Cases {
		if f.Name == "" || seen[f.Name] {
			t.Fatalf("empty or duplicate name %q", f.Name)
		}
		seen[f.Name] = true
		if f.Source == "" || (f.Expected == nil && !f.Unchanged) || (f.Unchanged && len(f.Problems) == 0) {
			t.Fatalf("fixture %s needs source and observable expectation", f.Name)
		}
		if f.Operation != "recipient" && f.Operation != "arrangement" && f.Operation != "remove" {
			t.Fatalf("fixture %s has unknown operation", f.Name)
		}
	}
	return m
}

func copyValue[T any](t *testing.T, value T) T {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var copied T
	if err := json.Unmarshal(data, &copied); err != nil {
		t.Fatal(err)
	}
	return copied
}

func call(editor contract.Editor, context contract.Context, current model.Preparation, f fixture) (model.Preparation, []model.Problem) {
	switch f.Operation {
	case "recipient":
		return editor.SetRecipient(context, current, f.Recipient)
	case "arrangement":
		return editor.SetArrangement(context, current, f.Arrangement)
	default:
		return editor.RemoveArrangement(context, current, f.ID)
	}
}

func TestFixtures(t *testing.T) {
	t.Parallel()
	m := readManifest(t)
	for _, f := range m.Cases {
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			context := copyValue(t, m.Context)
			if f.Context != nil {
				context = copyValue(t, *f.Context)
			}
			current := copyValue(t, m.Current)
			if f.Current != nil {
				current = copyValue(t, *f.Current)
			}
			f = copyValue(t, f)
			beforeContext, beforeCurrent, beforeFixture := copyValue(t, context), copyValue(t, current), copyValue(t, f)
			expected := f.Expected
			if f.Unchanged {
				expected = &beforeCurrent
			}
			editor := newEditor()
			got, problems := call(editor, context, current, f)
			if !reflect.DeepEqual(got, *expected) {
				t.Errorf("result %#v; want %#v", got, *expected)
			}
			checkProblems(t, problems, f.Problems)
			again, repeated := call(editor, context, current, f)
			if !reflect.DeepEqual(got, again) || !reflect.DeepEqual(problems, repeated) {
				t.Fatal("not deterministic")
			}
			if !reflect.DeepEqual(context, beforeContext) || !reflect.DeepEqual(current, beforeCurrent) || !reflect.DeepEqual(f, beforeFixture) {
				t.Fatal("input changed")
			}
		})
	}
}

func checkProblems(t *testing.T, got []model.Problem, expected []expectedProblem) {
	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf("problems %#v; want %#v", got, expected)
	}
	for i, p := range got {
		if p.Code != expected[i].Code || p.Detail["path"] != expected[i].Path || p.Severity != model.SeverityError || strings.TrimSpace(p.Message) == "" {
			t.Errorf("problem %d %#v; want %#v", i, p, expected[i])
		}
		for _, marker := range []string{"private-person@example.invalid", "another-person", "another-plan", "bad/account", "Bad ID", "no-longer-known", "New name must not leak"} {
			if strings.Contains(p.Message, marker) {
				t.Errorf("diagnostic exposes supplied data: %q", p.Message)
			}
		}
	}
}

func TestForkedEdits(t *testing.T) {
	t.Parallel()
	m := readManifest(t)
	current := copyValue(t, m.Current)
	current.Arrangements = append(make([]model.Arrangement, 0, 8), current.Arrangements...)
	editor := newEditor()
	first, firstProblems := editor.SetArrangement(m.Context, current, model.Arrangement{ID: "fork", Section: "section-one", Kind: "support", Status: "pending", Contacts: []string{"helper-one"}})
	checkProblems(t, firstProblems, nil)
	if len(first.Arrangements) != 3 {
		t.Fatalf("first edit arrangement count %d; want 3", len(first.Arrangements))
	}
	firstBefore := copyValue(t, first)
	second, secondProblems := editor.RemoveArrangement(m.Context, first, "z")
	checkProblems(t, secondProblems, nil)
	if len(second.Arrangements) != 2 || second.Arrangements[0].ID != "a" || second.Arrangements[1].ID != "fork" {
		t.Errorf("unexpected second edit %#v", second.Arrangements)
	}
	if !reflect.DeepEqual(first, firstBefore) || !reflect.DeepEqual(current, m.Current) {
		t.Fatal("later edit changed an earlier value")
	}
	third, thirdProblems := editor.SetArrangement(m.Context, current, model.Arrangement{ID: "branch", Section: "section-two", Kind: "support", Status: "deferred"})
	checkProblems(t, thirdProblems, nil)
	if len(third.Arrangements) != 3 || third.Arrangements[2].ID != "branch" {
		t.Errorf("unexpected fork %#v", third.Arrangements)
	}
	if !reflect.DeepEqual(first, firstBefore) {
		t.Fatal("fork changed earlier result")
	}
}

func TestOpaqueUnknownFields(t *testing.T) {
	t.Parallel()
	m := readManifest(t)
	opaque := &struct{ Label string }{Label: "preserved opaque data"}
	m.Current.UnknownFields["opaque"] = opaque
	got, problems := newEditor().SetRecipient(m.Context, m.Current, model.PreparationRecipient{Name: "After"})
	checkProblems(t, problems, nil)
	if !reflect.DeepEqual(got.UnknownFields["opaque"], &struct{ Label string }{Label: "preserved opaque data"}) {
		t.Fatal("edit converted or discarded opaque unknown data")
	}
	if opaque.Label != "preserved opaque data" || m.Current.Recipient.Name != "Before" || got.Recipient.Name != "After" {
		t.Fatal("edit changed an argument or did not apply the recipient")
	}
}
