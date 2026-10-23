package assembler_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	contract "onboard-kit/contracts/firststeps/assembler"
	"onboard-kit/internal/model"
)

type separatorCase struct {
	Name         string
	Source       string
	PlanURL      string
	ExpectedLine string
}

type pathExpectation struct {
	Path         string
	MinimumFence int
	Padding      bool
}

type fenceCase struct {
	Name        string
	Source      string
	ResultsPath string
	Record      pathExpectation
	Preparation pathExpectation
}

func readDisplayCases[T any](t *testing.T, path string) []T {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cases []T
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no independent display cases")
	}
	return cases
}

func displayInputs() (contract.Context, model.Plan, model.Preparation) {
	return contract.Context{Account: "newcomer"}, model.Plan{ID: "planone"}, model.Preparation{
		SchemaVersion: 1,
		Kind:          "preparation",
		Account:       "newcomer",
		PlanID:        "planone",
	}
}

func onlyDisplayLine(document, label string) (string, error) {
	var found []string
	for _, line := range strings.Split(document, "\n") {
		if strings.HasPrefix(line, label) {
			found = append(found, line)
		}
	}
	if len(found) != 1 {
		return "", fmt.Errorf("want exactly one %q line; got %d", label, len(found))
	}
	return found[0], nil
}

func literalPathSpan(line, label string, want pathExpectation) error {
	if !strings.HasPrefix(line, label) || !strings.HasSuffix(line, ".") {
		return fmt.Errorf("destination line has wrong label or final punctuation: %q", line)
	}
	span := strings.TrimSuffix(strings.TrimPrefix(line, label), ".")
	opening := 0
	for opening < len(span) && span[opening] == '`' {
		opening++
	}
	if opening < want.MinimumFence {
		return fmt.Errorf("opening width %d is below literal safe minimum %d", opening, want.MinimumFence)
	}
	closing := 0
	for closing < len(span) && span[len(span)-1-closing] == '`' {
		closing++
	}
	if opening != closing || len(span) < opening+closing {
		return fmt.Errorf("opening and closing fences differ or overlap: %q", span)
	}
	payload := span[opening : len(span)-closing]
	expected := want.Path
	if want.Padding {
		expected = " " + want.Path + " "
	}
	if payload != expected {
		return fmt.Errorf("literal path or required padding differs: got %q, want %q", payload, expected)
	}
	return nil
}

func TestWrapperLineSeparatorsFlattenIndependently(t *testing.T) {
	for _, item := range readDisplayCases[separatorCase](t, "testdata/line-separators.json") {
		t.Run(item.Name, func(t *testing.T) {
			t.Parallel()
			if item.Name == "" || item.Source == "" {
				t.Fatal("missing literal case provenance")
			}
			context, plan, preparation := displayInputs()
			plan.URL = item.PlanURL
			document, problems := newAssembler().Assemble(context, plan, preparation)
			if len(problems) != 0 {
				t.Fatalf("valid literal display input returned problems: %#v", problems)
			}
			line, err := onlyDisplayLine(document, "Plan location: ")
			if err != nil || line != item.ExpectedLine {
				t.Fatalf("flattened location got %q, want %q; %v", line, item.ExpectedLine, err)
			}
			if strings.ContainsAny(document, "\r\u2028\u2029") {
				t.Fatal("wrapper retained a line separator")
			}
			if plan.URL != item.PlanURL {
				t.Fatal("wrapper normalized the input location in place")
			}
		})
	}
}

func TestWrapperPathsUseSafeLiteralCodeSpans(t *testing.T) {
	for _, item := range readDisplayCases[fenceCase](t, "testdata/path-fences.json") {
		t.Run(item.Name, func(t *testing.T) {
			t.Parallel()
			if item.Name == "" || item.Source == "" {
				t.Fatal("missing literal case provenance")
			}
			context, plan, preparation := displayInputs()
			plan.ResultsPath = item.ResultsPath
			document, problems := newAssembler().Assemble(context, plan, preparation)
			if len(problems) != 0 {
				t.Fatalf("valid printable custom path returned problems: %#v", problems)
			}
			for _, expected := range []struct {
				label string
				value pathExpectation
			}{{"Relative record destination: ", item.Record}, {"Relative preparation destination: ", item.Preparation}} {
				line, err := onlyDisplayLine(document, expected.label)
				if err != nil {
					t.Fatal(err)
				}
				if err := literalPathSpan(line, expected.label, expected.value); err != nil {
					t.Fatal(err)
				}
			}
			if plan.ResultsPath != item.ResultsPath {
				t.Fatal("wrapper rewrote configured path in place")
			}
		})
	}
}

func TestPathOracleAllowsExtraWideSafeFences(t *testing.T) {
	for _, item := range []struct {
		line string
		want pathExpectation
	}{
		{"Path: ``````plain`path``````.", pathExpectation{Path: "plain`path", MinimumFence: 2}},
		{"Path: `````` `prefix ``````.", pathExpectation{Path: "`prefix", MinimumFence: 2, Padding: true}},
		{"Path: `````` suffix` ``````.", pathExpectation{Path: "suffix`", MinimumFence: 2, Padding: true}},
		{"Path: ``````pe```ople/re``co`rd``````.", pathExpectation{Path: "pe```ople/re``co`rd", MinimumFence: 4}},
	} {
		if err := literalPathSpan(item.line, "Path: ", item.want); err != nil {
			t.Fatal(err)
		}
	}
}
