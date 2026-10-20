package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

type screen int

const (
	screenQuestions screen = iota
	screenEdit
	screenSections
	screenReview
)

type App struct {
	backend       Backend
	audience      model.Audience
	save          func(model.Plan) error
	plan          model.Plan
	questions     []model.Question
	cursor        int
	screen        screen
	message       string
	edit          editState
	sectionCursor int
	review        viewport.Model
	width         int
	height        int
	finished      bool
}

type editState struct {
	question   model.Question
	choices    []choice
	choice     int
	input      textinput.Model
	suggestion int
	problems   []model.Problem
}

func NewApp(backend Backend, plan model.Plan, save func(model.Plan) error, audience model.Audience) App {
	app := App{backend: backend, save: save, plan: plan, audience: audience, width: 80, height: 24}
	app.questions = backend.Questions(plan, audience)
	return app
}

func (a App) Plan() model.Plan { return a.plan }

func (a App) Finished() bool { return a.finished }

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := message.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = typed.Width, typed.Height
		a.review.Width, a.review.Height = typed.Width, max(typed.Height-3, 3)
		return a, nil
	case tea.KeyMsg:
		if typed.String() == "ctrl+c" {
			a.finished = true
			return a, tea.Quit
		}
		switch a.screen {
		case screenEdit:
			return a.updateEdit(typed)
		case screenSections:
			return a.updateSections(typed)
		case screenReview:
			return a.updateReview(typed)
		}
		return a.updateQuestions(typed)
	}
	return a, nil
}

func (a App) updateQuestions(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q":
		a.finished = true
		return a, tea.Quit
	case "up", "k":
		a.cursor = max(a.cursor-1, 0)
	case "down", "j":
		a.cursor = min(a.cursor+1, max(len(a.questions)-1, 0))
	case "g", "home":
		a.cursor = 0
	case "G", "end":
		a.cursor = max(len(a.questions)-1, 0)
	case "enter":
		return a.startEdit(), nil
	case "d":
		return a.apply(model.Value{Kind: model.ValueKindDeferred}), nil
	case "x":
		return a.apply(model.Value{}), nil
	case "s":
		if a.audience == model.AudienceOnboardee {
			return a, nil
		}
		a.screen = screenSections
		a.message = ""
	case "r":
		return a.openReview(), nil
	}
	return a, nil
}

func (a App) current() (model.Question, bool) {
	if a.cursor < 0 || a.cursor >= len(a.questions) {
		return model.Question{}, false
	}
	return a.questions[a.cursor], true
}

func (a App) startEdit() App {
	question, ok := a.current()
	if !ok {
		return a
	}
	a.edit = editState{question: question}
	if picksFromList(question) {
		a.edit.choices = choicesFor(question)
		for index, option := range a.edit.choices {
			if sameValue(option.value, question.Answer) {
				a.edit.choice = index
			}
		}
	} else {
		input := textinput.New()
		input.SetValue(entryText(question))
		input.CursorEnd()
		input.Focus()
		input.Width = max(a.width-6, 20)
		a.edit.input = input
	}
	a.screen = screenEdit
	a.message = ""
	return a
}

func (a App) updateEdit(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.String() == "esc" {
		a.screen = screenQuestions
		return a, nil
	}
	if picksFromList(a.edit.question) {
		return a.updateChoice(key)
	}
	return a.updateText(key)
}

func (a App) updateChoice(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		a.edit.choice = max(a.edit.choice-1, 0)
	case "down", "j":
		a.edit.choice = min(a.edit.choice+1, len(a.edit.choices)-1)
	case "enter":
		return a.apply(a.edit.choices[a.edit.choice].value), nil
	}
	return a, nil
}

func (a App) updateText(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		return a.apply(parseEntry(a.edit.question, a.edit.input.Value())), nil
	case "ctrl+d":
		return a.apply(model.Value{Kind: model.ValueKindDeferred}), nil
	case "tab":
		return a.useSuggestion(), nil
	}
	var command tea.Cmd
	a.edit.input, command = a.edit.input.Update(key)
	return a, command
}

func (a App) useSuggestion() App {
	suggestions := a.edit.question.Suggestions
	if len(suggestions) == 0 {
		return a
	}
	suggestion := suggestions[a.edit.suggestion%len(suggestions)]
	a.edit.suggestion++
	a.edit.input.SetValue(describeValue(suggestion.Value))
	a.edit.input.CursorEnd()
	return a
}

func (a App) apply(value model.Value) App {
	question, ok := a.current()
	if !ok {
		return a
	}
	next, problems := a.backend.Submit(a.plan, question.ID, value)
	if hasError(problems) {
		a.edit.problems = problems
		a.message = firstMessage(problems)
		return a
	}
	return a.accept(next, question.ID, problems)
}

func (a App) accept(next model.Plan, keep string, problems []model.Problem) App {
	a.plan = next
	a.refresh(keep)
	a.screen = screenQuestions
	a.message = a.saveMessage(problems)
	return a
}

func (a App) saveMessage(problems []model.Problem) string {
	if err := a.save(a.plan); err != nil {
		return "could not save: " + err.Error()
	}
	if len(problems) > 0 {
		return "saved. warning: " + firstMessage(problems)
	}
	return "saved"
}

func (a *App) refresh(keep string) {
	a.questions = a.backend.Questions(a.plan, a.audience)
	a.cursor = min(a.cursor, max(len(a.questions)-1, 0))
	for index, question := range a.questions {
		if question.ID == keep {
			a.cursor = index
		}
	}
}

func (a App) updateSections(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc", "q":
		a.screen = screenQuestions
	case "up", "k":
		a.sectionCursor = max(a.sectionCursor-1, 0)
	case "down", "j":
		a.sectionCursor = min(a.sectionCursor+1, max(len(a.plan.Sections)-1, 0))
	case " ", "enter":
		return a.toggleSection(), nil
	}
	return a, nil
}

func (a App) toggleSection() App {
	if a.sectionCursor >= len(a.plan.Sections) {
		return a
	}
	section := a.plan.Sections[a.sectionCursor]
	next, problems := a.backend.SetSectionEnabled(a.plan, section.ID, !isOn(section))
	if hasError(problems) {
		a.message = firstMessage(problems)
		return a
	}
	a.plan = next
	a.refresh("")
	a.message = a.saveMessage(problems)
	return a
}

func (a App) openReview() App {
	document, _ := a.backend.Assemble(a.plan)
	a.review = viewport.New(a.width, max(a.height-3, 3))
	a.review.SetContent(a.reviewText(document))
	a.screen = screenReview
	return a
}

func (a App) reviewText(document string) string {
	var text strings.Builder
	blockers := a.backend.Blockers(a.plan, a.audience)
	text.WriteString(blockerSummary(blockers))
	problems := a.backend.Validate(a.plan)
	for _, problem := range problems {
		text.WriteString(problem.Question + " " + problem.Code + ": " + problem.Message + "\n")
	}
	text.WriteString("\n" + document)
	return text.String()
}

func (a App) updateReview(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc", "q":
		a.screen = screenQuestions
		return a, nil
	}
	var command tea.Cmd
	a.review, command = a.review.Update(key)
	return a, command
}

func isOn(section model.Section) bool {
	return section.Enabled == nil || *section.Enabled
}

func hasError(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func firstMessage(problems []model.Problem) string {
	if len(problems) == 0 {
		return ""
	}
	return problems[0].Message
}

func blockerSummary(blockers []model.Blocker) string {
	if len(blockers) == 0 {
		return "No required question is waiting for an answer.\n"
	}
	var text strings.Builder
	text.WriteString("Required questions with no answer:\n")
	for _, blocker := range blockers {
		text.WriteString("  - " + blocker.Question + "\n")
	}
	return text.String()
}

func sameValue(left model.Value, right model.Value) bool {
	return left.Kind == right.Kind && describeValue(left) == describeValue(right)
}
