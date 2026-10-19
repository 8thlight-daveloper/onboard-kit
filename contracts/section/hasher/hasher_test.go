package hasher_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type fixtureCase struct {
	Name      string   `json:"name"`
	Relation  string   `json:"relation"`
	Inputs    []string `json:"inputs"`
	Canonical string   `json:"canonical"`
	Hash      string   `json:"hash"`
	Source    string   `json:"source"`
}

var hashShape = regexp.MustCompile(`^[0-9a-f]{64}$`)

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

func loadCaseSections(t *testing.T, c fixtureCase) []model.Section {
	t.Helper()
	var sections []model.Section
	for _, input := range c.Inputs {
		sections = append(sections, loadSections(t, input)...)
	}
	return sections
}

func loadOneSection(t *testing.T, name string) model.Section {
	t.Helper()
	sections := loadSections(t, name)
	if len(sections) != 1 {
		t.Fatalf("%s: want one section, got %d", name, len(sections))
	}
	return sections[0]
}

func textHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func TestEveryCaseCitesItsSourceAndHasAKnownShape(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
		switch c.Relation {
		case "golden":
			if len(c.Inputs) != 1 || c.Canonical == "" || !hashShape.MatchString(c.Hash) {
				t.Errorf("case %s: a golden case has one input, a canonical text, and a hash", c.Name)
			}
		case "same":
			if len(c.Inputs) < 2 || !hashShape.MatchString(c.Hash) {
				t.Errorf("case %s: a same case has two inputs or more, and the hash they share", c.Name)
			}
		case "different":
			if len(loadCaseSections(t, c)) < 2 {
				t.Errorf("case %s: a different case has two sections or more", c.Name)
			}
		case "every-field":
			if len(c.Inputs) != 1 {
				t.Errorf("case %s: an every-field case has one input", c.Name)
			}
		default:
			t.Errorf("case %s has the unknown relation %q", c.Name, c.Relation)
		}
	}
}

func TestEveryGoldenHashIsTheSha256OfItsCanonicalText(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Relation != "golden" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got := textHash(string(readFixture(t, c.Canonical)))
			if got != c.Hash {
				t.Errorf("the hash in cases.json is %s, the canonical text hashes to %s", c.Hash, got)
			}
		})
	}
}

func TestHashesTheGoldenSections(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Relation != "golden" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got := newHasher().Hash(loadOneSection(t, c.Inputs[0]))
			if got != c.Hash {
				t.Errorf("got %s, want %s\ncanonical text: %s", got, c.Hash, readFixture(t, c.Canonical))
			}
		})
	}
}

func TestGivesOneHashToSectionsWithTheSameContent(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Relation != "same" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			hasher := newHasher()
			for i, section := range loadCaseSections(t, c) {
				if got := hasher.Hash(section); got != c.Hash {
					t.Errorf("section %d: got %s, want %s\n%s", i, got, c.Hash, sectionText(t, section))
				}
			}
		})
	}
}

func TestGivesEverySectionWithDifferentContentItsOwnHash(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Relation != "different" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			hasher := newHasher()
			seen := map[string]int{}
			for i, section := range loadCaseSections(t, c) {
				got := hasher.Hash(section)
				if first, ok := seen[got]; ok {
					t.Errorf("sections %d and %d share the hash %s\n%s\n%s", first, i, got, sectionText(t, loadCaseSections(t, c)[first]), sectionText(t, section))
				}
				seen[got] = i
			}
		})
	}
}

func sectionText(t *testing.T, section model.Section) string {
	t.Helper()
	data, err := jsonform.SectionJSON(section)
	if err != nil {
		return fmt.Sprintf("%+v", section)
	}
	return string(data)
}

var proseFields = map[string]bool{
	"Title": true, "Summary": true, "Body": true, "Verification": true,
	"Notes": true, "Description": true, "PatternHint": true, "Ask": true,
}

type change struct {
	path  string
	apply func()
}

func changesTo(v reflect.Value, path string) []change {
	var changes []change
	if _, isValue := v.Interface().(model.Value); isValue {
		return valueChanges(v, path)
	}
	switch v.Kind() {
	case reflect.String:
		changes = append(changes, change{path, func() { v.SetString(v.String() + "x") }})
		if !proseFields[path[strings.LastIndex(path, ".")+1:]] {
			changes = append(changes, change{path + " with a trailing space", func() { v.SetString(v.String() + " ") }})
		}
	case reflect.Bool:
		changes = append(changes, change{path, func() { v.SetBool(!v.Bool()) }})
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		changes = append(changes, change{path + " removed", func() { v.Set(reflect.Zero(v.Type())) }})
		if v.Elem().Kind() == reflect.Bool {
			changes = append(changes, change{path, func() { v.Elem().SetBool(!v.Elem().Bool()) }})
			return changes
		}
		changes = append(changes, changesTo(v.Elem(), path)...)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			name := v.Type().Field(i).Name
			if name == "Source" {
				continue
			}
			changes = append(changes, changesTo(v.Field(i), path+"."+name)...)
		}
	case reflect.Slice:
		changes = append(changes, change{path + " appended", func() { v.Set(reflect.Append(v, reflect.Zero(v.Type().Elem()))) }})
		for i := 0; i < v.Len(); i++ {
			changes = append(changes, changesTo(v.Index(i), fmt.Sprintf("%s[%d]", path, i))...)
		}
	case reflect.Map:
		changes = append(changes, change{path + " added", func() { v.SetMapIndex(reflect.ValueOf("zz-added"), mapAddition(v.Type().Elem())) }})
		for _, key := range v.MapKeys() {
			changes = append(changes, change{path + "." + key.String(), func() { v.SetMapIndex(key, mapReplacement(v.MapIndex(key))) }})
			if v.MapIndex(key).Kind() == reflect.String {
				changes = append(changes, change{path + "." + key.String() + " with a trailing space", func() { v.SetMapIndex(key, reflect.ValueOf(v.MapIndex(key).String()+" ")) }})
			}
		}
	default:
		panic(fmt.Sprintf("%s: the test does not know the kind %s", path, v.Kind()))
	}
	return changes
}

func mapAddition(elem reflect.Type) reflect.Value {
	if elem.Kind() == reflect.Interface {
		return reflect.ValueOf("added")
	}
	return reflect.ValueOf("added").Convert(elem)
}

func mapReplacement(current reflect.Value) reflect.Value {
	if current.Kind() == reflect.Interface {
		return reflect.ValueOf(fmt.Sprintf("changed-%v", current.Interface()))
	}
	return reflect.ValueOf(current.String() + "x")
}

func valueChanges(v reflect.Value, path string) []change {
	value := v.Interface().(model.Value)
	var changes []change
	switch value.Kind {
	case model.ValueKindText:
		changes = append(changes, change{path, func() { v.Set(reflect.ValueOf(model.Value{Kind: model.ValueKindText, Text: value.Text + "x"})) }})
		changes = append(changes, change{path + " with a trailing space", func() { v.Set(reflect.ValueOf(model.Value{Kind: model.ValueKindText, Text: value.Text + " "})) }})
	case model.ValueKindFlag:
		changes = append(changes, change{path, func() { v.Set(reflect.ValueOf(model.Value{Kind: model.ValueKindFlag, Flag: !value.Flag})) }})
	case model.ValueKindList:
		changes = append(changes, change{path, func() {
			v.Set(reflect.ValueOf(model.Value{Kind: model.ValueKindList, List: append(append([]string{}, value.List...), "x")}))
		}})
	}
	changes = append(changes, change{path + " as text", func() { v.Set(reflect.ValueOf(model.Value{Kind: model.ValueKindText, Text: "changed"})) }})
	return changes
}

func TestEveryFieldChangesTheHash(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Relation != "every-field" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			hasher := newHasher()
			base := hasher.Hash(loadOneSection(t, c.Inputs[0]))
			probe := loadOneSection(t, c.Inputs[0])
			count := len(changesTo(reflect.ValueOf(&probe).Elem(), "Section"))
			if count < 60 {
				t.Fatalf("only %d changes found, the walk is missing fields", count)
			}
			for i := 0; i < count; i++ {
				section := loadOneSection(t, c.Inputs[0])
				chosen := changesTo(reflect.ValueOf(&section).Elem(), "Section")[i]
				chosen.apply()
				if got := hasher.Hash(section); got == base {
					t.Errorf("a change at %s did not change the hash", chosen.path)
				}
			}
		})
	}
}

func TestEverySectionGetsALowercaseSha256HexHash(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		for _, section := range loadCaseSections(t, c) {
			if got := newHasher().Hash(section); !hashShape.MatchString(got) {
				t.Errorf("%s: %q is not 64 lowercase hex digits", c.Name, got)
			}
		}
	}
}

func TestAnEmptySectionHashesOnlyItsDefaultTiming(t *testing.T) {
	t.Parallel()
	if got, want := newHasher().Hash(model.Section{}), textHash(`{"timing":"anytime"}`); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestEmptyAndAbsentListsAndMapsHashAlike(t *testing.T) {
	t.Parallel()
	absent := model.Section{ID: "s", Steps: []model.Step{{ID: "a"}}, Variables: []model.Variable{{Name: "v"}}}
	empty := model.Section{
		ID:            "s",
		After:         []string{},
		Escalation:    []string{},
		Variables:     []model.Variable{{Name: "v", Options: []string{}}},
		Steps:         []model.Step{{ID: "a", Escalation: []string{}, Run: []string{}, Env: map[string]string{}, Checks: []model.Check{}}},
		UnknownFields: map[string]any{},
	}
	if got, want := newHasher().Hash(empty), newHasher().Hash(absent); got != want {
		t.Errorf("empty lists and maps hash as %s, absent ones as %s", got, want)
	}
	noItems := model.Section{ID: "s", Variables: []model.Variable{}, Steps: []model.Step{}}
	if got, want := newHasher().Hash(noItems), textHash(`{"id":"s","timing":"anytime"}`); got != want {
		t.Errorf("empty lists of items hash as %s, want %s", got, want)
	}
}

func TestEmptyConditionsHashAsAbsentOnes(t *testing.T) {
	t.Parallel()
	absent := model.Section{ID: "s", Variables: []model.Variable{{Name: "v"}}}
	empty := model.Section{ID: "s", IncludeIf: &model.Condition{All: []model.Clause{}}, Variables: []model.Variable{{Name: "v", AskIf: &model.Condition{}}}}
	want := textHash(`{"id":"s","timing":"anytime","variables":[{"name":"v","type":"string","phase":"planning"}]}`)
	for name, section := range map[string]model.Section{"absent": absent, "empty": empty} {
		if got := newHasher().Hash(section); got != want {
			t.Errorf("%s conditions: got %s, want %s", name, got, want)
		}
	}
}

func TestTheModelsDefaultsHashAsAbsentFields(t *testing.T) {
	t.Parallel()
	absent := model.Section{ID: "s", Variables: []model.Variable{{Name: "v"}}}
	explicit := model.Section{ID: "s", Timing: model.DefaultTiming, Variables: []model.Variable{{Name: "v", Type: model.DefaultVariableType, Phase: model.DefaultPhase}}}
	want := textHash(`{"id":"s","timing":"anytime","variables":[{"name":"v","type":"string","phase":"planning"}]}`)
	for name, section := range map[string]model.Section{"absent": absent, "explicit": explicit} {
		if got := newHasher().Hash(section); got != want {
			t.Errorf("%s defaults: got %s, want %s", name, got, want)
		}
	}
}

func TestIntegersOfEveryWidthHashAlikeAndDifferFromDecimals(t *testing.T) {
	t.Parallel()
	hasherOf := func(value any) string {
		return newHasher().Hash(model.Section{ID: "s", UnknownFields: map[string]any{"x": value}})
	}
	if hasherOf(int(5)) != hasherOf(int64(5)) {
		t.Errorf("int and int64 with the same number hash differently")
	}
	if hasherOf(int(5)) != textHash(`{"id":"s","timing":"anytime","unknownFields":{"x":5}}`) {
		t.Errorf("an integer is not written in plain decimal")
	}
	if hasherOf(float64(5)) != textHash(`{"id":"s","timing":"anytime","unknownFields":{"x":5.0}}`) {
		t.Errorf("a whole decimal is not written with .0")
	}
	if hasherOf(float64(1.25)) != textHash(`{"id":"s","timing":"anytime","unknownFields":{"x":1.25}}`) {
		t.Errorf("a decimal is not written in plain notation")
	}
}

func TestAValueIsHashedByItsKindAlone(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		value model.Value
		want  string
	}{
		"text with stray fields":     {model.Value{Kind: model.ValueKindText, Text: "a", Flag: true, List: []string{"z"}}, `,"default":"a"`},
		"flag with stray fields":     {model.Value{Kind: model.ValueKindFlag, Flag: true, Text: "z", List: []string{"z"}}, `,"default":true`},
		"false flag":                 {model.Value{Kind: model.ValueKindFlag, Text: "z"}, `,"default":false`},
		"list with stray fields":     {model.Value{Kind: model.ValueKindList, List: []string{"a"}, Text: "z", Flag: true}, `,"default":["a"]`},
		"deferred with stray fields": {model.Value{Kind: model.ValueKindDeferred, Text: "z", Flag: true, List: []string{"z"}}, `,"default":{"deferred":true}`},
		"none with stray fields":     {model.Value{Text: "z", Flag: true, List: []string{"z"}}, ``},
		"nil list":                   {model.Value{Kind: model.ValueKindList}, `,"default":[]`},
		"empty list":                 {model.Value{Kind: model.ValueKindList, List: []string{}}, `,"default":[]`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			section := model.Section{ID: "s", Variables: []model.Variable{{Name: "v", Default: c.value}}}
			want := textHash(`{"id":"s","timing":"anytime","variables":[{"name":"v","type":"string"` + c.want + `,"phase":"planning"}]}`)
			if got := newHasher().Hash(section); got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}

func TestEveryCaseHashesTheSameOnEveryCall(t *testing.T) {
	t.Parallel()
	shared := newHasher()
	for _, c := range loadCases(t) {
		for i, section := range loadCaseSections(t, c) {
			first := newHasher().Hash(section)
			for call := 0; call < 8; call++ {
				if again := newHasher().Hash(section); again != first {
					t.Fatalf("%s[%d]: two calls gave %s and %s", c.Name, i, first, again)
				}
				if again := shared.Hash(section); again != first {
					t.Fatalf("%s[%d]: a reused hasher gave %s, a new one %s", c.Name, i, again, first)
				}
			}
		}
	}
}

func TestTheSectionIsNotChanged(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		sections := loadCaseSections(t, c)
		snapshots := loadCaseSections(t, c)
		hasher := newHasher()
		for i, section := range sections {
			hasher.Hash(section)
			if !reflect.DeepEqual(section, snapshots[i]) {
				t.Errorf("%s[%d]: the section changed during Hash\n%s\n%s", c.Name, i, sectionText(t, section), sectionText(t, snapshots[i]))
			}
		}
	}
}
