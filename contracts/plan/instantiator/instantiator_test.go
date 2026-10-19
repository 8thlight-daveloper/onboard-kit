package instantiator_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type fixtureSection struct {
	ID        string `json:"id"`
	Hash      string `json:"hash"`
	Canonical string `json:"canonical"`
}

type fixtureProblem struct {
	Code    string            `json:"code"`
	Section string            `json:"section"`
	Detail  map[string]string `json:"detail"`
}

type fixtureCase struct {
	Name          string           `json:"name"`
	Templates     []string         `json:"templates"`
	AxisVariables string           `json:"axisVariables"`
	TemplateSet   string           `json:"templateSet"`
	GeneratedOn   string           `json:"generatedOn"`
	Sections      []fixtureSection `json:"sections"`
	Problems      []fixtureProblem `json:"problems"`
	Source        string           `json:"source"`
}

type inputs struct {
	Templates     []model.Section
	AxisVariables []model.Variable
}

type outcome struct {
	Plan     model.Plan
	Problems []model.Problem
}

var hashShape = regexp.MustCompile(`^[0-9a-f]{64}$`)

const supportedSchemaVersion = 1

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

func loadCases(t *testing.T) []fixtureCase {
	t.Helper()
	var cases []fixtureCase
	if err := json.Unmarshal(readFixture(t, "cases.json"), &cases); err != nil {
		t.Fatalf("cases.json: %v", err)
	}
	return cases
}

func loadSections(t *testing.T, name string) []model.Section {
	t.Helper()
	data := bytes.TrimSpace(readFixture(t, name))
	var documents []json.RawMessage
	if bytes.HasPrefix(data, []byte("[")) {
		if err := json.Unmarshal(data, &documents); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	} else {
		documents = []json.RawMessage{data}
	}
	var sections []model.Section
	for i, document := range documents {
		section, err := jsonform.Section(document)
		if err != nil {
			t.Fatalf("%s[%d]: %v", name, i, err)
		}
		sections = append(sections, section)
	}
	return sections
}

func loadInputs(t *testing.T, c fixtureCase) inputs {
	t.Helper()
	var in inputs
	for _, name := range c.Templates {
		in.Templates = append(in.Templates, loadSections(t, name)...)
	}
	if c.AxisVariables != "" {
		plan, err := jsonform.Plan(readFixture(t, c.AxisVariables))
		if err != nil {
			t.Fatalf("%s: %v", c.AxisVariables, err)
		}
		in.AxisVariables = plan.Variables
	}
	return in
}

func (c fixtureCase) expectsPlan() bool {
	return len(c.Problems) == 0
}

func run(instantiator interface {
	Instantiate([]model.Section, []model.Variable, string, model.Date) (model.Plan, []model.Problem)
}, c fixtureCase, in inputs) outcome {
	plan, problems := instantiator.Instantiate(in.Templates, in.AxisVariables, c.TemplateSet, model.Date(c.GeneratedOn))
	return outcome{plan, problems}
}

func normalized(plan model.Plan) model.Plan {
	return withoutEmptyCollections(reflect.ValueOf(plan)).Interface().(model.Plan)
}

func withoutEmptyCollections(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Slice:
		if v.Len() == 0 {
			return reflect.Zero(v.Type())
		}
		copied := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			copied.Index(i).Set(withoutEmptyCollections(v.Index(i)))
		}
		return copied
	case reflect.Map:
		if v.Len() == 0 {
			return reflect.Zero(v.Type())
		}
		copied := reflect.MakeMapWithSize(v.Type(), v.Len())
		for _, key := range v.MapKeys() {
			copied.SetMapIndex(key, withoutEmptyCollections(v.MapIndex(key)))
		}
		return copied
	case reflect.Pointer:
		if v.IsNil() {
			return v
		}
		copied := reflect.New(v.Type().Elem())
		copied.Elem().Set(withoutEmptyCollections(v.Elem()))
		return copied
	case reflect.Interface:
		if v.IsNil() {
			return v
		}
		copied := reflect.New(v.Type()).Elem()
		copied.Set(withoutEmptyCollections(v.Elem()))
		return copied
	case reflect.Struct:
		copied := reflect.New(v.Type()).Elem()
		for i := 0; i < v.NumField(); i++ {
			copied.Field(i).Set(withoutEmptyCollections(v.Field(i)))
		}
		return copied
	default:
		return v
	}
}

func planText(t *testing.T, plan model.Plan) string {
	t.Helper()
	data, err := jsonform.JSON(normalized(plan))
	if err != nil {
		t.Fatalf("write the plan: %v", err)
	}
	return string(data)
}

func expectedPlan(t *testing.T, c fixtureCase, in inputs) model.Plan {
	t.Helper()
	byID := map[string]model.Section{}
	for _, template := range in.Templates {
		byID[template.ID] = template
	}
	plan := model.Plan{
		SchemaVersion: supportedSchemaVersion,
		Provenance:    model.Provenance{TemplateSet: c.TemplateSet, GeneratedOn: model.Date(c.GeneratedOn)},
		Variables:     in.AxisVariables,
	}
	for _, want := range c.Sections {
		section, ok := byID[want.ID]
		if !ok {
			t.Fatalf("case %s: no template has the id %s", c.Name, want.ID)
		}
		section.Source = &model.Source{Template: want.ID, Hash: want.Hash}
		plan.Sections = append(plan.Sections, section)
	}
	return plan
}

func textHash(text []byte) string {
	sum := sha256.Sum256(text)
	return hex.EncodeToString(sum[:])
}

func TestEveryCaseCitesItsSourceAndHasAKnownShape(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
		if c.expectsPlan() {
			in := loadInputs(t, c)
			plan := expectedPlan(t, c, in)
			if len(plan.Sections) != len(in.Templates) {
				t.Errorf("case %s: %d sections expected for %d templates", c.Name, len(plan.Sections), len(in.Templates))
			}
			for _, want := range c.Sections {
				if !hashShape.MatchString(want.Hash) || want.Canonical == "" {
					t.Errorf("case %s, section %s: a section has a hash and the canonical text it comes from", c.Name, want.ID)
				}
			}
			continue
		}
		if len(c.Sections) != 0 {
			t.Errorf("case %s: a case with problems has no plan", c.Name)
		}
		for _, problem := range c.Problems {
			if problem.Code == "" {
				t.Errorf("case %s: a problem has no code", c.Name)
			}
		}
	}
}

func TestEveryHashIsTheSha256OfItsCanonicalText(t *testing.T) {
	for _, c := range loadCases(t) {
		for _, want := range c.Sections {
			t.Run(c.Name+"/"+want.ID, func(t *testing.T) {
				t.Parallel()
				got := textHash(readFixture(t, want.Canonical))
				if got != want.Hash {
					t.Errorf("the hash in cases.json is %s, the canonical text hashes to %s", want.Hash, got)
				}
			})
		}
	}
}

func TestBuildsTheDefaultPlan(t *testing.T) {
	for _, c := range loadCases(t) {
		if !c.expectsPlan() {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			in := loadInputs(t, c)
			want := expectedPlan(t, c, loadInputs(t, c))
			got := run(newInstantiator(), c, in)
			if len(got.Problems) != 0 {
				t.Fatalf("want no problems, got %+v", got.Problems)
			}
			if got.Plan.SchemaVersion != supportedSchemaVersion {
				t.Errorf("schemaVersion: got %d, want %d", got.Plan.SchemaVersion, supportedSchemaVersion)
			}
			if len(got.Plan.Answers.Plan) != 0 || len(got.Plan.Answers.Sections) != 0 {
				t.Errorf("want no answers, got %+v", got.Plan.Answers)
			}
			if len(got.Plan.UnknownFields) != 0 {
				t.Errorf("want no unknown fields on the plan, got %v", got.Plan.UnknownFields)
			}
			if g, w := planText(t, got.Plan), planText(t, want); g != w {
				t.Errorf("the plan differs\ngot:\n%s\nwant:\n%s", g, w)
			}
		})
	}
}

func TestReportsEveryProblemAndReturnsNoPlan(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.expectsPlan() {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got := run(newInstantiator(), c, loadInputs(t, c))
			if !reflect.DeepEqual(normalized(got.Plan), model.Plan{}) {
				t.Errorf("want the zero plan, got\n%s", planText(t, got.Plan))
			}
			if len(got.Problems) != len(c.Problems) {
				t.Fatalf("got %d problems, want %d: %+v", len(got.Problems), len(c.Problems), got.Problems)
			}
			for i, want := range c.Problems {
				problem := got.Problems[i]
				if problem.Code != want.Code || problem.Section != want.Section || problem.Question != "" {
					t.Errorf("problem %d: got code %q, section %q, question %q, want code %q, section %q", i, problem.Code, problem.Section, problem.Question, want.Code, want.Section)
				}
				if problem.Severity != model.SeverityError {
					t.Errorf("problem %d: severity %q, want error", i, problem.Severity)
				}
				if strings.TrimSpace(problem.Message) == "" {
					t.Errorf("problem %d: no message", i)
				}
				if !reflect.DeepEqual(problem.Detail, want.Detail) {
					t.Errorf("problem %d: detail %v, want %v", i, problem.Detail, want.Detail)
				}
			}
		})
	}
}

func TestResultsDoNotDependOnEarlierCalls(t *testing.T) {
	t.Parallel()
	cases := loadCases(t)
	first := make([]outcome, len(cases))
	for i, c := range cases {
		first[i] = run(newInstantiator(), c, loadInputs(t, c))
	}
	shared := newInstantiator()
	for pass := 0; pass < 2; pass++ {
		for i := range cases {
			index := i
			if pass == 1 {
				index = len(cases) - 1 - i
			}
			got := run(shared, cases[index], loadInputs(t, cases[index]))
			if !reflect.DeepEqual(got, first[index]) {
				t.Fatalf("case %s on pass %d: the result differs from a new instantiator's", cases[index].Name, pass)
			}
		}
	}
}

func TestOrderOfTemplatesDoesNotChangeTheResult(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		in := loadInputs(t, c)
		want := run(newInstantiator(), c, loadInputs(t, c))
		reversed := inputs{AxisVariables: in.AxisVariables}
		for i := len(in.Templates) - 1; i >= 0; i-- {
			reversed.Templates = append(reversed.Templates, in.Templates[i])
		}
		got := run(newInstantiator(), c, reversed)
		if planText(t, got.Plan) != planText(t, want.Plan) || !reflect.DeepEqual(got.Problems, want.Problems) {
			t.Errorf("case %s: reversing the templates changed the result", c.Name)
		}
	}
}

func TestInputsAreNotChanged(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		in := loadInputs(t, c)
		run(newInstantiator(), c, in)
		if want := loadInputs(t, c); !reflect.DeepEqual(in, want) {
			t.Errorf("case %s: an argument was changed", c.Name)
		}
	}
}

func scribble(v reflect.Value) {
	switch v.Kind() {
	case reflect.String:
		v.SetString(v.String() + "!")
	case reflect.Bool:
		v.SetBool(!v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(v.Int() + 1)
	case reflect.Float32, reflect.Float64:
		v.SetFloat(v.Float() + 1)
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			scribble(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			element := reflect.New(v.Type().Elem()).Elem()
			element.Set(v.MapIndex(key))
			scribble(element)
			v.SetMapIndex(key, element)
		}
	case reflect.Pointer:
		if !v.IsNil() {
			scribble(v.Elem())
		}
	case reflect.Interface:
		if !v.IsNil() {
			element := reflect.New(v.Elem().Type()).Elem()
			element.Set(v.Elem())
			scribble(element)
			v.Set(element)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			scribble(v.Field(i))
		}
	}
}

func TestResultSharesNothingWithTheArguments(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if !c.expectsPlan() {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			in := loadInputs(t, c)
			got := run(newInstantiator(), c, in)
			want := run(newInstantiator(), c, loadInputs(t, c))
			scribble(reflect.ValueOf(&in).Elem())
			if !reflect.DeepEqual(got, want) {
				t.Errorf("changing the arguments after the call changed the result\ngot:\n%s\nwant:\n%s", planText(t, got.Plan), planText(t, want.Plan))
			}

			in = loadInputs(t, c)
			got = run(newInstantiator(), c, in)
			scribble(reflect.ValueOf(&got).Elem())
			if untouched := loadInputs(t, c); !reflect.DeepEqual(in, untouched) {
				t.Errorf("changing the result changed an argument")
			}
		})
	}
}
