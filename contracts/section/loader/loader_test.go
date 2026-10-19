package loader_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code   string            `json:"code"`
	Path   string            `json:"path"`
	Detail map[string]string `json:"detail"`
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
	var cases []fixtureCase
	if err := json.Unmarshal(readFixture(t, "cases.json"), &cases); err != nil {
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

func loadSection(t *testing.T, name string) model.Section {
	t.Helper()
	section, err := jsonform.Section(readFixture(t, name))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return section
}

func describe(section model.Section) string {
	written, err := jsonform.SectionJSON(section)
	if err != nil {
		return err.Error()
	}
	return string(written)
}

func actualProblems(problems []model.Problem) []expectedProblem {
	var actual []expectedProblem
	for _, p := range problems {
		detail := map[string]string{}
		for key, value := range p.Detail {
			if key != "path" {
				detail[key] = value
			}
		}
		if len(detail) == 0 {
			detail = nil
		}
		actual = append(actual, expectedProblem{Code: p.Code, Path: p.Detail["path"], Detail: detail})
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

func TestEveryCaseExpectsASectionOrAProblem(t *testing.T) {
	for _, c := range loadCases(t) {
		if (c.Output == "") == (len(c.Problems) == 0) {
			t.Errorf("case %s needs an output or at least one problem, and not both", c.Name)
		}
	}
}

func TestLoadsTheTemplate(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output == "" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newLoader().Load(readFixture(t, c.Input))
			if len(problems) != 0 {
				t.Fatalf("unexpected problems: %+v", actualProblems(problems))
			}
			want := loadSection(t, c.Output)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("sections differ\n--- got ---\n%s\n--- want ---\n%s", describe(got), describe(want))
			}
		})
	}
}

func TestReportsProblemsAndReturnsNoSection(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output != "" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newLoader().Load(readFixture(t, c.Input))
			if !reflect.DeepEqual(got, model.Section{}) {
				t.Errorf("want the zero section, got\n%s", describe(got))
			}
			for _, p := range problems {
				if p.Severity != model.SeverityError {
					t.Errorf("problem %s has severity %q, want error", p.Code, p.Severity)
				}
				if strings.TrimSpace(p.Message) == "" {
					t.Errorf("problem %s at %q has no message", p.Code, p.Detail["path"])
				}
			}
			if actual := actualProblems(problems); !reflect.DeepEqual(actual, c.Problems) {
				t.Errorf("problems differ\n got: %+v\nwant: %+v", actual, c.Problems)
			}
		})
	}
}

func TestLoadedSectionHasNoSourceNoUnknownFieldsAndNoSwitch(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output == "" {
			continue
		}
		got, _ := newLoader().Load(readFixture(t, c.Input))
		if got.Source != nil || got.UnknownFields != nil || got.Enabled != nil {
			t.Errorf("%s: want no source, no unknown fields, and no switch, got\n%s", c.Name, describe(got))
		}
	}
}

func TestResultIsDeterministic(t *testing.T) {
	for _, c := range loadCases(t) {
		data := readFixture(t, c.Input)
		firstSection, firstProblems := newLoader().Load(data)
		for i := 0; i < 5; i++ {
			section, problems := newLoader().Load(data)
			if !reflect.DeepEqual(firstSection, section) || !reflect.DeepEqual(firstProblems, problems) {
				t.Fatalf("%s: two calls gave different results", c.Name)
			}
		}
	}
}

func TestInputIsNotChanged(t *testing.T) {
	byteOrderMark := []byte("\xef\xbb\xbf")
	for _, c := range loadCases(t) {
		file := readFixture(t, c.Input)
		for name, data := range map[string][]byte{"as is": file, "with a byte order mark": append(bytes.Clone(byteOrderMark), file...)} {
			snapshot := bytes.Clone(data)
			newLoader().Load(data)
			if !bytes.Equal(data, snapshot) {
				t.Errorf("%s %s: the bytes changed during Load", c.Name, name)
			}
		}
	}
}

func TestResultsShareNothingBetweenCalls(t *testing.T) {
	data := readFixture(t, "all-fields.yaml")
	want := loadSection(t, "all-fields.section.json")
	first, _ := newLoader().Load(data)
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("first result differs\n%s", describe(first))
	}
	first.After[0] = "changed"
	first.Escalation[0] = "changed"
	first.IncludeIf.All[0].Question = "changed"
	first.IncludeIf.All[0].Answer.IsOneOf[0].Text = "changed"
	first.Variables[1].Options[0] = "changed"
	first.Variables[3].Default.List[0] = "changed"
	first.Variables[4].AskIf.All[0].Question = "changed"
	first.Variables[4].AskIf.All[2].Answer.IsOneOf[0].Text = "changed"
	first.Steps[0].Escalation[0] = "changed"
	first.Steps[0].Run[0] = "changed"
	first.Steps[0].Env["APP_ENV"] = "changed"
	first.Steps[0].Checks[0].Run[0] = "changed"
	second, _ := newLoader().Load(data)
	if !reflect.DeepEqual(second, want) {
		t.Errorf("a change to one result showed in the next\n%s", describe(second))
	}
}

func TestReadsWindowsLineEndings(t *testing.T) {
	data := "schemaVersion: 1\r\nid: s\r\ntitle: T\r\nsummary: S\r\nbody: |\r\n  One\r\n  Two\r\nsteps:\r\n  - id: st\r\n    name: St\r\n"
	input := []byte(data)
	got, problems := newLoader().Load(input)
	want := loadSection(t, "windows-line-endings.section.json")
	if string(input) != data {
		t.Errorf("the bytes changed during Load")
	}
	if len(problems) != 0 || !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v and\n%s\nwant\n%s", actualProblems(problems), describe(got), describe(want))
	}
}

func TestReadsAByteOrderMark(t *testing.T) {
	data := append([]byte("\xef\xbb\xbf"), readFixture(t, "minimal.yaml")...)
	got, problems := newLoader().Load(data)
	if len(problems) != 0 || !reflect.DeepEqual(got, loadSection(t, "minimal.section.json")) {
		t.Errorf("got %+v and\n%s", actualProblems(problems), describe(got))
	}
}
