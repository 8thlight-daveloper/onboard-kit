package tui

import (
	"fmt"
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

type stagedChange struct {
	text    string
	details []string
}

type preparationReview struct {
	changes       []stagedChange
	problems      []model.Problem
	blockers      []model.Blocker
	guide         string
	guideProblems []model.Problem
	reviewed      []byte
	refusal       string
	focusGuide    bool
	scroll        int
	guideScroll   int
}

func reviewedPerson(person openPerson) preparationReview {
	review := preparationReview{changes: stagedChanges(person)}
	result := reviewOf(person.loaded, person.draft)
	review.problems, review.blockers = result.PreparationProblems, result.OnboarderBlockers
	review.guide, review.guideProblems = result.FirstSteps, result.FirstStepsProblems
	if person.loaded.Exists && !person.staged() {
		review.reviewed = person.loaded.Bytes
		return review
	}
	reviewed, problems := encoded(person.draft)
	switch {
	case hasError(problems):
		review.refusal = "This preparation cannot be written as a file: " + firstMessage(problems)
	case len(reviewed) > preparationByteLimit:
		review.refusal = "This preparation is larger than 1 MiB, so it cannot be saved."
	default:
		review.reviewed = reviewed
	}
	return review
}

func stagedChanges(person openPerson) []stagedChange {
	plan := person.loaded.Plan
	held, draft := person.loaded.Preparation, person.draft
	changes := fieldChanges("", recipientSpecs(), held.Recipient, draft.Recipient, plan)
	for _, arrangement := range held.Arrangements {
		if indexOfArrangement(draft.Arrangements, arrangement.ID) < 0 {
			changes = append(changes, stagedChange{text: "Arrangement " + arrangementID(arrangement) + ": removed"})
		}
	}
	for _, arrangement := range draft.Arrangements {
		label := "Arrangement " + arrangementID(arrangement)
		index := indexOfArrangement(held.Arrangements, arrangement.ID)
		if index >= 0 {
			changes = append(changes, fieldChanges(label+", ", arrangementSpecs(), held.Arrangements[index], arrangement, plan)...)
			continue
		}
		changes = append(changes, stagedChange{text: label + ": added", details: addedFields(arrangement, plan)})
	}
	return changes
}

func indexOfArrangement(arrangements []model.Arrangement, id string) int {
	return slices.IndexFunc(arrangements, func(each model.Arrangement) bool { return each.ID == id })
}

func fieldChanges[T any](prefix string, specs []fieldSpec[T], was T, now T, plan model.Plan) []stagedChange {
	var changes []stagedChange
	for _, spec := range specs {
		before, after := spec.read(was), spec.read(now)
		if !slices.Equal(before, after) {
			changes = append(changes, stagedChange{text: prefix + spec.label + ": " + quotedValues(spec.kind, after, plan) + ", was " + quotedValues(spec.kind, before, plan)})
		}
	}
	return changes
}

func addedFields(arrangement model.Arrangement, plan model.Plan) []string {
	var details []string
	for _, spec := range arrangementSpecs() {
		if values := spec.read(arrangement); !slices.Equal(values, spec.read(model.Arrangement{})) {
			details = append(details, spec.label+": "+quotedValues(spec.kind, values, plan))
		}
	}
	return details
}

func (app App) openPreparationReview() App {
	app.preparation.review = reviewedPerson(app.preparation.person)
	app.preparation.mode = preparationReviewing
	return app
}

func (app App) savedPreparation() App {
	review := app.preparation.review
	if review.refusal != "" {
		app.status = line{text: "not saved: " + review.refusal, tone: toneProblem}
		return app
	}
	person := app.preparation.person
	outcome := app.preparationPort.Save(person.loaded.Account, review.reviewed, person.loaded.FoldersMissing)
	switch {
	case outcome.committed():
		return app.afterSave(outcome)
	case outcome == PreparationUnchanged:
		app.preparation.mode = preparationPerson
		app.status = line{text: outcome.sentence(), tone: toneFaint}
		return app
	case outcome == PreparationStale:
		return app.reappliedAfterStale()
	}
	app.status = line{text: "not saved: " + outcome.sentence(), tone: toneProblem}
	return app
}

func (app App) afterSave(outcome PreparationOutcome) App {
	destination := app.preparation.person.loaded.Destination
	loaded, reloaded := app.preparationPort.Load(app.preparation.person.loaded.Account)
	if reloaded == PreparationDone {
		app.preparation.person = newPerson(loaded)
		app.preparation.mode = preparationPerson
		app.preparation.row = min(app.preparation.row, app.preparation.person.rowCount()-1)
	} else {
		app = app.closedPerson()
	}
	app.status = line{text: "saved " + destination, tone: toneFaint}
	if outcome == PreparationSavedLeftovers {
		app.status = line{text: outcome.sentence(), tone: toneWarning}
	}
	return app
}

func (app App) reappliedAfterStale() App {
	person := app.preparation.person
	loaded, outcome := app.preparationPort.Load(person.loaded.Account)
	if outcome != PreparationDone {
		app.status = line{text: "not saved: " + PreparationStale.sentence() + " " + outcome.sentence(), tone: toneProblem}
		return app
	}
	fresh, dropped := person.reapplied(loaded)
	app.preparation.person = fresh
	app = app.openPreparationReview()
	text := "Nothing was saved, because the files changed on disk. Your changes are applied to the new version. Review them, then press y."
	if dropped > 0 {
		text = fmt.Sprintf("Nothing was saved, because the files changed on disk. %s no longer %s and %s dropped. Review the rest, then press y.", pluralWords(dropped, "change", "changes"), pluralVerb(dropped, "applies", "apply"), pluralVerb(dropped, "was", "were"))
	}
	app.status = line{text: text, tone: toneWarning}
	return app
}

func pluralVerb(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func (app App) blockerTexts(blockers []model.Blocker) []string {
	var texts []string
	for _, blocker := range blockers {
		text := blocker.Question
		if index := app.indexOf(blocker.Question); index >= 0 {
			text = questionTitle(app.questions[index], blocker.Question)
		}
		texts = append(texts, "required: "+text)
	}
	return texts
}

func problemTexts(problems []model.Problem) []string {
	var texts []string
	for _, problem := range problems {
		texts = append(texts, string(problem.Severity)+": "+problem.Message)
	}
	return texts
}

func guideLines(guide string) []string {
	var shown []string
	for _, source := range strings.Split(guide, "\n") {
		shown = append(shown, visibleText(strings.ReplaceAll(source, "\t", "    ")))
	}
	return shown
}
