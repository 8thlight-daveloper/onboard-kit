package tui

import (
	"reflect"
	"slices"

	"github.com/charmbracelet/bubbles/textinput"

	"onboard-kit/internal/model"
)

type preparationMode int

const (
	preparationListing preparationMode = iota
	preparationAdding
	preparationPerson
	preparationEditing
	preparationRemoving
	preparationDropping
	preparationReviewing
)

type changeKind int

const (
	changeRecipient changeKind = iota
	changeArrangement
	changeRemoval
)

type preparationChange struct {
	kind        changeKind
	recipient   model.PreparationRecipient
	arrangement model.Arrangement
}

type openPerson struct {
	loaded  PreparationLoad
	draft   model.Preparation
	changes []preparationChange
}

type preparationState struct {
	mode           preparationMode
	people         []PreparationPerson
	listRefusal    PreparationOutcome
	cursor         int
	account        textinput.Model
	accountProblem string
	person         openPerson
	row            int
	form           preparationForm
	review         preparationReview
}

func (app App) WithPreparation(port PreparationPort) App {
	app.preparationPort = port
	return app
}

func (app App) canPrepare() bool {
	return app.preparationPort != nil && app.audience != model.AudienceOnboardee
}

func newPerson(loaded PreparationLoad) openPerson {
	return openPerson{loaded: loaded, draft: loaded.Preparation}
}

func (person openPerson) staged() bool {
	return !reflect.DeepEqual(person.loaded.Preparation, person.draft)
}

func (person openPerson) withChange(change preparationChange) (openPerson, []model.Problem) {
	next, problems := change.applied(person.loaded, person.draft)
	if hasError(problems) {
		return person, problems
	}
	person.draft = next
	person.changes = append(slices.Clone(person.changes), change)
	return person, nil
}

func (person openPerson) reapplied(loaded PreparationLoad) (openPerson, int) {
	fresh := newPerson(loaded)
	dropped := 0
	for _, change := range person.changes {
		next, problems := fresh.withChange(change)
		if problems != nil {
			dropped++
			continue
		}
		fresh = next
	}
	return fresh, dropped
}

func (person openPerson) arrangements() []model.Arrangement { return person.draft.Arrangements }

func (person openPerson) rowCount() int { return len(person.arrangements()) + 3 }

func (person openPerson) arrangementRow(row int) (int, bool) {
	index := row - 1
	return index, index >= 0 && index < len(person.arrangements())
}

func (person openPerson) addRow() int { return len(person.arrangements()) + 1 }

func (person openPerson) reviewRow() int { return len(person.arrangements()) + 2 }

func (person openPerson) sharedIDs() []string {
	var shared []string
	seen := map[string]bool{}
	for _, arrangement := range person.draft.Arrangements {
		if seen[arrangement.ID] && !slices.Contains(shared, arrangement.ID) {
			shared = append(shared, arrangement.ID)
		}
		seen[arrangement.ID] = true
	}
	return shared
}

func (app App) openPreparation() App {
	if !app.canPrepare() {
		return app
	}
	app = app.leavingForSideScreen()
	app.screen = screenPreparation
	switch app.preparation.mode {
	case preparationListing:
		return app.listedPeople()
	case preparationPerson:
		if !app.preparation.person.staged() && app.preparation.person.loaded.Exists {
			return app.reloadedPerson()
		}
	}
	return app
}

func (app App) listedPeople() App {
	people, outcome := app.preparationPort.People()
	app.preparation.people = people
	app.preparation.listRefusal = outcome
	app.preparation.cursor = max(min(app.preparation.cursor, len(people)), 0)
	return app
}

func (app App) reloadedPerson() App {
	loaded, outcome := app.preparationPort.Load(app.preparation.person.loaded.Account)
	if outcome != PreparationDone {
		return app
	}
	app.preparation.person = newPerson(loaded)
	app.preparation.row = min(app.preparation.row, app.preparation.person.rowCount()-1)
	return app
}

func (app App) openedPerson(account string) (App, PreparationOutcome) {
	loaded, outcome := app.preparationPort.Load(account)
	if outcome != PreparationDone {
		return app, outcome
	}
	app.preparation.person = newPerson(loaded)
	app.preparation.mode = preparationPerson
	app.preparation.row = 0
	return app, outcome
}

func (app App) closedPerson() App {
	app.preparation.person = openPerson{}
	app.preparation.mode = preparationListing
	return app.listedPeople()
}

func (app App) preparationTyping() bool {
	switch app.preparation.mode {
	case preparationAdding:
		return true
	case preparationEditing:
		return !app.preparation.form.choosing && app.preparation.form.onText()
	}
	return false
}
