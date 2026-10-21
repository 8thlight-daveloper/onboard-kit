package tui

import (
	"strings"

	"onboard-kit/internal/model"
)

type fakeBackend struct {
	questions []model.Question
	answers   map[string]model.Value
	submitted []string
}

func newFakeBackend(questions ...model.Question) *fakeBackend {
	return &fakeBackend{questions: questions, answers: map[string]model.Value{}}
}

func (backend *fakeBackend) Questions(model.Plan, model.Audience) []model.Question {
	asked := make([]model.Question, 0, len(backend.questions))
	for _, question := range backend.questions {
		question.Answer = backend.answers[question.ID]
		asked = append(asked, question)
	}
	return asked
}

func (backend *fakeBackend) Blockers(model.Plan, model.Audience) []model.Blocker { return nil }

func (backend *fakeBackend) Submit(plan model.Plan, question string, value model.Value) (model.Plan, []model.Problem) {
	backend.submitted = append(backend.submitted, question)
	backend.answers[question] = value
	return plan, nil
}

func (backend *fakeBackend) SetSectionEnabled(plan model.Plan, _ string, _ bool) (model.Plan, []model.Problem) {
	return plan, nil
}

func (backend *fakeBackend) Validate(model.Plan) []model.Problem { return nil }

func (backend *fakeBackend) Assemble(model.Plan) (string, []model.Problem) {
	return "# Document\n\nA line of the document.", nil
}

func textQuestion(id string, scope string, description string, example string) model.Question {
	return model.Question{ID: id, Scope: scope, Name: id, Description: description, Example: example, Type: model.VariableTypeString, Required: true, Audience: model.AudienceOnboarder}
}

func directoryQuestion(id string, scope string, description string, variableType model.VariableType, multiple bool, choices ...model.Choice) model.Question {
	question := textQuestion(id, scope, description, "")
	question.Type = variableType
	question.Multiple = multiple
	question.Choices = choices
	return question
}

func people() []model.Choice {
	return []model.Choice{{Value: "alex-rivera", Label: "Alex Rivera"}, {Value: "sam-okafor", Label: "Sam Okafor"}, {Value: "it-help", Label: "IT service desk"}}
}

func startFake(backend *fakeBackend) App {
	return startFakeIn(backend, model.Plan{}, model.AudienceOnboarder)
}

func startFakeIn(backend *fakeBackend, plan model.Plan, audience model.Audience) App {
	saved := func(model.Plan) error { return nil }
	app := NewApp(backend, plan, saved, audience)
	app.rail = unicodeRailGlyphs()
	return app
}

func sectionsWithParents(pairs ...string) model.Plan {
	var plan model.Plan
	for index := 0; index+1 < len(pairs); index += 2 {
		plan.Sections = append(plan.Sections, model.Section{ID: pairs[index], Title: strings.ReplaceAll(strings.ToUpper(pairs[index][:1])+pairs[index][1:], "-", " "), Parent: pairs[index+1]})
	}
	return plan
}

func longScreens() map[string]App {
	long := strings.Repeat("A very long description that keeps going and going. ", 12)
	var questions []model.Question
	for index := 0; index < 40; index++ {
		scope := []string{"plan", "access", "tools", "people"}[index%4]
		question := textQuestion(scope+".q"+strings.Repeat("x", index), scope, long, strings.Repeat("an example ", 20))
		question.Suggestions = []model.Suggestion{{Value: model.Value{Kind: model.ValueKindText, Text: long}, Source: model.SuggestionSourceDefault}}
		questions = append(questions, question)
	}
	walking := startFake(newFakeBackend(questions...))
	var crowd []model.Choice
	for index := 0; index < 40; index++ {
		crowd = append(crowd, model.Choice{Value: "id" + strings.Repeat("x", index), Label: long[:30+index]})
	}
	picking := startFake(newFakeBackend(directoryQuestion("plan.crowd", "plan", long, model.VariableTypeContact, true, crowd...)))
	return map[string]App{
		"long walk":     walking,
		"long overview": press(walking, esc),
		"long review":   press(walking, esc, typed("r")),
		"long picks":    press(picking, typed("j"), space, typed("j"), space),
	}
}
