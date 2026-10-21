package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

const (
	startWidth  = 120
	startHeight = 30
)

type App struct {
	backend        Backend
	rail           railGlyphs
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
	contacts       contactsState
	width          int
	height         int
	finished       bool
}

func NewApp(backend Backend, plan model.Plan, save func(model.Plan) error, audience model.Audience) App {
	app := App{rail: railGlyphsForTerminal(), backend: backend, save: save, plan: plan, audience: audience, width: startWidth, height: startHeight}
	app.questions = backend.Questions(plan, audience)
	return app.refreshedContacts().opening()
}

func (app App) WithNotice(text string) App {
	if text != "" {
		app.status = line{text: text, tone: toneStrong}
	}
	return app
}

func (app App) opening() App {
	if len(app.questions) == 0 || app.holdsAnswers() {
		return app.openReview()
	}
	app.returnTo = screenOverview
	return app.openQuestion(0)
}

func (app App) Plan() model.Plan { return app.plan }

func (app App) Finished() bool { return app.finished }

func (app App) Init() tea.Cmd { return nil }

func (app App) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := message.(type) {
	case tea.WindowSizeMsg:
		app.width, app.height = typed.Width, typed.Height
		return app.resized(), nil
	case tea.KeyMsg:
		return app.updateKey(typed)
	}
	return app, nil
}

func (app App) frame() frame { return newFrame(app.width, app.height) }

func (app App) resized() App {
	app.entry = app.entry.withWidth(app.inputWidth())
	app.review = app.review.scrolled(app.documentOffsetLimit(), 0)
	if app.contacts.mode == contactsEditing {
		app.contacts.form = app.contacts.form.withWidth(app.formWidth())
	}
	return app
}

func (app App) updateKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.String() == "ctrl+c" {
		return app.quit()
	}
	app.status = line{}
	if app.helpOpen {
		app.helpOpen = false
		return app, nil
	}
	switch app.screen {
	case screenOverview:
		return app.updateOverview(key)
	case screenReview:
		return app.updateReview(key)
	case screenSections:
		return app.updateSections(key)
	case screenContacts:
		return app.updateContacts(key)
	}
	return app.updateWalk(key)
}

func (app App) quit() (tea.Model, tea.Cmd) {
	app.finished = true
	return app, tea.Quit
}
