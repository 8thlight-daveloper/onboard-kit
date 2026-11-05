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
	backend         Backend
	rail            railGlyphs
	audience        model.Audience
	save            func(model.Plan) error
	plan            model.Plan
	questions       []model.Question
	cursor          int
	overviewCursor  int
	sectionCursor   int
	screen          screen
	returnTo        screen
	cameFrom        screen
	helpOpen        bool
	stripFocused    bool
	stripChoice     int
	about           *About
	aboutOpen       bool
	aboutScroll     int
	status          line
	entry           entryState
	review          reviewState
	contacts        contactsState
	preparation     preparationState
	preparationPort PreparationPort
	invitation      invitationState
	invitationPort  InvitationPort
	width           int
	height          int
	finished        bool
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

func (app App) WithAbout(about *About) App {
	app.about = about
	return app
}

func (app App) updateAbout(key tea.KeyMsg) App {
	switch key.String() {
	case "esc", "q", "a", "enter", "?":
		app.aboutOpen = false
		return app
	}
	height := app.frame().bodyHeight
	if offset, moved := scrolledBy(key.String(), app.aboutScroll, height); moved {
		content := aboutLines(*app.about, app.aboutWidth())
		app.aboutScroll = max(min(offset, aboutScrollLimit(content, height)), 0)
	}
	return app
}

func (app App) aboutWidth() int {
	return min(app.frame().innerWidth(), readingWidth)
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
	if app.preparation.mode == preparationEditing {
		app.preparation.form = app.preparation.form.withWidth(app.formWidth())
	}
	if app.invitation.mode == invitationEditing || app.invitation.mode == invitationConfirming {
		app.invitation.form = app.invitation.form.withWidth(app.invitationFormWidth())
	}
	return app
}

func (app App) updateKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.String() == "ctrl+c" {
		return app.quit()
	}
	app.status = line{}
	if app.aboutOpen {
		return app.updateAbout(key), nil
	}
	if app.helpOpen {
		app.helpOpen = false
		if key.String() == "a" && app.about != nil {
			app.aboutOpen, app.aboutScroll = true, 0
		}
		return app, nil
	}
	if app.stripFocused {
		return app.updateStrip(key)
	}
	if step := pageStep(key.String()); step != 0 && app.tabMovesPages() {
		return app.pageMoved(step), nil
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
	case screenPreparation:
		return app.updatePreparation(key)
	case screenInvitation:
		return app.updateInvitation(key)
	}
	return app.updateAnswer(key)
}

func (app App) quit() (tea.Model, tea.Cmd) {
	app.finished = true
	return app, tea.Quit
}
