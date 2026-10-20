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

func startFake(backend *fakeBackend) App {
	saved := func(model.Plan) error { return nil }
	app := NewApp(backend, model.Plan{}, saved, model.AudienceOnboarder)
	app.rail = unicodeRailGlyphs()
	return app
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
	return map[string]App{
		"long walk":     walking,
		"long overview": press(walking, esc),
		"long review":   press(walking, esc, typed("r")),
	}
}
