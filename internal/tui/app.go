package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

const (
	startWidth  = 80
	startHeight = 24
)

type App struct {
	backend        Backend
	audience       model.Audience
	save           func(model.Plan) error
	plan           model.Plan
	questions      []model.Question
	cursor         int
	overviewCursor int
	sectionCursor  int
	screen         screen
	returnTo       screen
	cameFrom       screen
	helpOpen       bool
	status         line
	entry          entryState
	review         reviewState
	width          int
	height         int
	finished       bool
}

func NewApp(backend Backend, plan model.Plan, save func(model.Plan) error, audience model.Audience) App {
	app := App{backend: backend, save: save, plan: plan, audience: audience, width: startWidth, height: startHeight}
	app.questions = backend.Questions(plan, audience)
	return app.opening()
}

func (a App) opening() App {
	if len(a.questions) == 0 || a.holdsAnswers() {
		return a.openReview()
	}
	a.returnTo = screenOverview
	return a.openQuestion(0)
}

func (a App) holdsAnswers() bool {
	for _, question := range a.questions {
		if stateOf(question.Answer) != "unanswered" {
			return true
		}
	}
	return false
}

func (a App) Plan() model.Plan { return a.plan }

func (a App) Finished() bool { return a.finished }

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := message.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = typed.Width, typed.Height
		return a.resized(), nil
	case tea.KeyMsg:
		return a.updateKey(typed)
	}
	return a, nil
}

func (a App) resized() App {
	a.entry = a.entry.withWidth(layoutFor(a.width, a.height).mainWidth)
	a.review = a.review.scrolled(a.documentOffsetLimit(), 0)
	return a
}

func (a App) updateKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.String() == "ctrl+c" {
		return a.quit()
	}
	a.status = line{}
	if a.helpOpen {
		a.helpOpen = false
		return a, nil
	}
	switch a.screen {
	case screenOverview:
		return a.updateOverview(key)
	case screenReview:
		return a.updateReview(key)
	case screenSections:
		return a.updateSections(key)
	}
	return a.updateWalk(key)
}

func (a App) quit() (tea.Model, tea.Cmd) {
	a.finished = true
	return a, tea.Quit
}

func (a App) current() (model.Question, bool) {
	if a.cursor < 0 || a.cursor >= len(a.questions) {
		return model.Question{}, false
	}
	return a.questions[a.cursor], true
}

func (a App) canSwitchSections() bool { return a.audience != model.AudienceOnboardee }

func (a App) openQuestion(index int) App {
	a.cursor = max(min(index, len(a.questions)-1), 0)
	if question, ok := a.current(); ok {
		a.entry = newEntryState(question, layoutFor(a.width, a.height).mainWidth)
	}
	a.screen = screenWalk
	return a
}

func (a App) openOverview() App {
	a.overviewCursor = a.cursor
	a.screen = screenOverview
	a.returnTo = screenOverview
	return a
}

func (a App) openReview() App {
	document, _ := a.backend.Assemble(a.plan)
	a.review = newReviewState(a.backend, a.plan, a.audience, a.questions, document)
	a.screen = screenReview
	a.returnTo = screenReview
	return a
}

func (a App) openSections() App {
	if !a.canSwitchSections() {
		return a
	}
	a.cameFrom = a.screen
	a.screen = screenSections
	return a
}

func (a App) closeSections() App {
	switch a.cameFrom {
	case screenOverview:
		a.screen = screenOverview
		a.overviewCursor = min(a.overviewCursor, max(len(a.questions)-1, 0))
		return a
	case screenReview:
		return a.openReview()
	}
	return a.openQuestion(a.cursor)
}

func (a App) returnToWalk() App {
	if len(a.questions) == 0 {
		return a
	}
	return a.openQuestion(a.cursor)
}

func (a App) moveTo(index int) App {
	if index >= len(a.questions) {
		return a.openReview()
	}
	return a.openQuestion(max(index, 0))
}

func (a App) jumpTo(index int) App {
	if index < 0 || index >= len(a.questions) {
		return a
	}
	return a.openQuestion(index)
}

func (a App) firstUnanswered() int {
	for index, question := range a.questions {
		if stateOf(question.Answer) == "unanswered" {
			return index
		}
	}
	return a.cursor
}

func (a App) indexOf(id string) int {
	for index, question := range a.questions {
		if question.ID == id {
			return index
		}
	}
	return -1
}

func (a App) submit(value model.Value) (App, bool) {
	question, ok := a.current()
	if !ok {
		return a, false
	}
	next, problems := a.backend.Submit(a.plan, question.ID, value)
	if hasError(problems) {
		a.entry.problems = problems
		a.status = line{text: "not saved", tone: toneProblem}
		return a, false
	}
	a.plan = next
	a.refresh(question.ID)
	a.status = a.saveStatus(problems)
	return a, true
}

func (a App) saveStatus(problems []model.Problem) line {
	if err := a.save(a.plan); err != nil {
		return line{text: "could not save: " + err.Error(), tone: toneProblem}
	}
	if len(problems) > 0 {
		return line{text: "saved. warning: " + firstMessage(problems), tone: toneWarning}
	}
	return line{text: "saved", tone: toneMuted}
}

func (a *App) refresh(keep string) {
	a.questions = a.backend.Questions(a.plan, a.audience)
	a.cursor = min(a.cursor, max(len(a.questions)-1, 0))
	if index := a.indexOf(keep); index >= 0 {
		a.cursor = index
	}
	a.overviewCursor = min(a.overviewCursor, max(len(a.questions)-1, 0))
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
