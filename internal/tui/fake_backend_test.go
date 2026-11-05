package tui

import (
	"strings"

	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/editor"
)

type fakeBackend struct {
	questions []model.Question
	answers   map[string]model.Value
	askIf     map[string]func(map[string]model.Value) bool
	submitted []string
}

func newFakeBackend(questions ...model.Question) *fakeBackend {
	return &fakeBackend{questions: questions, answers: map[string]model.Value{}}
}

func (backend *fakeBackend) Questions(model.Plan, model.Audience) []model.Question {
	asked := make([]model.Question, 0, len(backend.questions))
	for _, question := range backend.questions {
		if asked, conditional := backend.askIf[question.ID]; conditional && !asked(backend.answers) {
			continue
		}
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

func (backend *fakeBackend) SetContact(plan model.Plan, contact model.Contact) (model.Plan, []model.Problem) {
	return editor.New().SetContact(plan, contact)
}

func (backend *fakeBackend) SetRole(plan model.Plan, role model.Role) (model.Plan, []model.Problem) {
	return editor.New().SetRole(plan, role)
}

func (backend *fakeBackend) SetTeam(plan model.Plan, team model.Team) (model.Plan, []model.Problem) {
	return editor.New().SetTeam(plan, team)
}

func (backend *fakeBackend) RemoveContact(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem) {
	return editor.New().RemoveContact(plan, id)
}

func (backend *fakeBackend) RemoveRole(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem) {
	return editor.New().RemoveRole(plan, id)
}

func (backend *fakeBackend) RemoveTeam(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem) {
	return editor.New().RemoveTeam(plan, id)
}

func (backend *fakeBackend) PruneHiddenAnswers(plan model.Plan) (model.Plan, model.Cleared, []model.Problem) {
	return editor.New().PruneHiddenAnswers(plan)
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
	answering := startFake(newFakeBackend(questions...))
	var crowd []model.Choice
	for index := 0; index < 40; index++ {
		crowd = append(crowd, model.Choice{Value: "id" + strings.Repeat("x", index), Label: long[:30+index]})
	}
	picking := startFake(newFakeBackend(directoryQuestion("plan.crowd", "plan", long, model.VariableTypeContact, true, crowd...)))
	return map[string]App{
		"long contacts": press(crowdedContacts(), typed("c")),
		"long roles":    press(crowdedContacts(), typed("c"), right),
		"long teams":    press(crowdedContacts(), typed("c"), right, right),
		"long holders":  press(crowdedContacts(), typed("c"), right, enter, enter, down, space, down, space),
		"long override": press(crowdedContacts(), typed("c"), right, right, enter, enter, enter, down, down, space),
		"long contact":  press(crowdedContacts(), typed("c"), enter, typed("more text that goes on and on and on and on and on and on and on and on and on")),
		"long removal":  press(crowdedContacts(), typed("c"), typed("x")),
		"long answer":   answering,
		"long blurred":  press(answering, typed(long), esc, down),
		"long strip":    press(answering, up, right),
		"long typed":    press(answering, typed(long)),
		"long overview": press(answering, esc, esc),
		"long review":   press(answering, esc, typed("r")),
		"long picks":    press(picking, typed("j"), space, typed("j"), space),
		"long page":     longPage().openQuestion(7),
		"long page top": longPage(),
		"long page off": press(longPage().openQuestion(5), esc, up),
	}
}

func longPage() App {
	long := strings.Repeat("A long question that keeps going. ", 6)
	var questions []model.Question
	for index := 0; index < 10; index++ {
		question := textQuestion("plan.row"+strings.Repeat("x", index), "plan", long, strings.Repeat("an example ", 12))
		if index%3 == 1 {
			question = directoryQuestion("plan.pick"+strings.Repeat("x", index), "plan", long, model.VariableTypeContact, index%2 == 0, people()...)
		}
		questions = append(questions, question)
	}
	return startFake(newFakeBackend(questions...))
}

func crowdedContacts() App {
	long := strings.Repeat("A very long name that keeps going ", 4)
	var plan model.Plan
	var ids []string
	for index := 0; index < 40; index++ {
		id := "contact" + strings.Repeat("x", index)
		ids = append(ids, id)
		plan.Directory.Contacts = append(plan.Directory.Contacts, model.Contact{ID: id, Name: long[:20+index], Email: "someone" + strings.Repeat("long", index) + "@example.test", Slack: "@" + id, SlackID: "U" + id})
	}
	plan.Directory.Roles = []model.Role{{ID: "lead", Name: long, Holders: ids}, {ID: "second-role-with-a-long-id", Name: long[:30]}}
	plan.Directory.Teams = []model.Team{{ID: "crowd", Name: long, Holders: map[string][]string{"lead": ids, "second-role-with-a-long-id": {}}}}
	plan.Variables = []model.Variable{{Name: "who", Description: long, Type: model.VariableTypeContact, Multiple: true}}
	plan.Answers = model.Answers{Plan: map[string]model.Value{"who": {Kind: model.ValueKindList, List: ids}}}
	backend := newFakeBackend(textQuestion("plan.who", "plan", long, ""))
	return resize(startFakeIn(backend, plan, model.AudienceOnboarder).openOverview(), 120, 30)
}
