package resolver_test

import (
	"slices"
	"testing"

	"onboard-kit/internal/model"
)

func TestOnlyStoredAnswersResolveRegardlessOfSuggestions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		kind        model.VariableType
		answer      model.Value
		replacement model.Value
		text        string
	}{
		{"implicit-string", "", model.Value{Kind: model.ValueKindText, Text: "accepted"}, model.Value{Kind: model.ValueKindText, Text: "changed"}, "accepted"},
		{"string", model.VariableTypeString, model.Value{Kind: model.ValueKindText, Text: "accepted"}, model.Value{Kind: model.ValueKindText, Text: "changed"}, "accepted"},
		{"choice", model.VariableTypeChoice, model.Value{Kind: model.ValueKindText, Text: "accepted"}, model.Value{Kind: model.ValueKindText, Text: "changed"}, "accepted"},
		{"boolean", model.VariableTypeBoolean, model.Value{Kind: model.ValueKindFlag, Flag: false}, model.Value{Kind: model.ValueKindFlag, Flag: true}, "false"},
		{"command", model.VariableTypeCommand, model.Value{Kind: model.ValueKindList, List: []string{"make", "test"}}, model.Value{Kind: model.ValueKindList, List: []string{"go", "test"}}, "make test"},
		{"contact", model.VariableTypeContact, model.Value{Kind: model.ValueKindText, Text: "sam-okafor"}, model.Value{Kind: model.ValueKindText, Text: "lee-park"}, "Sam Okafor, @sam"},
		{"role", model.VariableTypeRole, model.Value{Kind: model.ValueKindText, Text: "it-contact"}, model.Value{Kind: model.ValueKindText, Text: "mentor"}, "IT help desk, #it-help"},
		{"team", model.VariableTypeTeam, model.Value{Kind: model.ValueKindText, Text: "platform"}, model.Value{Kind: model.ValueKindText, Text: "design"}, "Platform team"},
	}
	for _, c := range cases {
		for _, audience := range []model.Audience{"", model.AudienceOnboarder, model.AudienceOnboardee} {
			t.Run(c.name+"/"+string(audience), func(t *testing.T) {
				t.Parallel()
				states := []struct {
					name   string
					answer model.Value
					text   string
					code   string
				}{
					{"accepted", c.answer, c.text, ""},
					{"missing", model.Value{}, "_unanswered_", "gap.unanswered"},
					{"deferred", model.Value{Kind: model.ValueKindDeferred}, "_deferred_", "gap.deferred"},
				}
				if audience == model.AudienceOnboardee {
					states[1].text = "<value>"
				}
				if c.kind == model.VariableTypeContact || c.kind == model.VariableTypeRole || c.kind == model.VariableTypeTeam {
					states = append(states, struct {
						name   string
						answer model.Value
						text   string
						code   string
					}{"unknown-id", model.Value{Kind: model.ValueKindText, Text: "unknown-id"}, states[1].text, "gap.unanswered"})
				}
				for _, state := range states {
					for _, suggestion := range []model.Value{c.answer, c.replacement, {}, {Kind: model.ValueKindDeferred}} {
						section := model.Section{ID: "suggestions", Body: "$value", Variables: []model.Variable{{Name: "value", Type: c.kind, Audience: audience, Default: suggestion}}}
						answers := model.Answers{Sections: map[string]map[string]model.Value{"suggestions": {"value": state.answer}}}
						got, problems := newResolver().Resolve(section, answers, loadDirectory(t, "directory.plan.json"), "")
						if got.Body != state.text {
							t.Errorf("%s with suggestion %+v: body %q, want %q", state.name, suggestion, got.Body, state.text)
						}
						if state.code == "" {
							if len(problems) != 0 {
								t.Errorf("accepted answer: unexpected problems %+v", problems)
							}
							continue
						}
						wantAudience := audience
						if wantAudience == "" {
							wantAudience = model.AudienceOnboarder
						}
						if len(problems) != 1 || problems[0].Code != state.code || problems[0].Question != "suggestions.value" || problems[0].Section != section.ID || problems[0].Severity != model.SeverityWarning || problems[0].Detail["audience"] != string(wantAudience) {
							t.Errorf("%s: problems %+v, want %s at suggestions.value for %s", state.name, problems, state.code, wantAudience)
						}
					}
				}
			})
		}
	}
}

func TestShortRunsAreLiteralAndThreeCharacterNamesResolve(t *testing.T) {
	t.Parallel()
	text := "$x $ab $a0 $a_ $abc $a09 $a_0 $abcx $abcA $$abc"
	section := model.Section{
		ID:   "names",
		Body: text,
		Variables: []model.Variable{
			{Name: "abc", Type: model.VariableTypeCommand},
			{Name: "a09", Type: model.VariableTypeCommand},
			{Name: "a_0", Type: model.VariableTypeCommand},
		},
		Steps: []model.Step{{Run: []string{"$x", "$ab", "$a0", "$a_", "$abc", "$a09", "$a_0", text}}},
	}
	answers := model.Answers{Sections: map[string]map[string]model.Value{"names": {
		"x":   {Kind: model.ValueKindText, Text: "not-declared"},
		"ab":  {Kind: model.ValueKindText, Text: "not-declared"},
		"abc": {Kind: model.ValueKindList, List: []string{"make", "test"}},
		"a09": {Kind: model.ValueKindList, List: []string{"go", "test"}},
		"a_0": {Kind: model.ValueKindList, List: []string{"go", "build"}},
	}}}
	got, problems := newResolver().Resolve(section, answers, model.Directory{}, "")
	wantBody := "$x $ab $a0 $a_ make test go test go build $abcx make testA $abc"
	if got.Body != wantBody {
		t.Errorf("body %q", got.Body)
	}
	wantRun := []string{"$x", "$ab", "$a0", "$a_", "make", "test", "go", "test", "go", "build", wantBody}
	if !slices.Equal(got.Steps[0].Run, wantRun) {
		t.Errorf("run %q, want %q", got.Steps[0].Run, wantRun)
	}
	if len(problems) != 0 {
		t.Errorf("unexpected problems %+v", problems)
	}
}
