package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

const (
	shownValueCharacters  = 20
	hiddenValuePrefix     = 4
	leftOverShownLimit    = 5
	nothingLeftOverNotice = "No answers are left over from removed questions."
)

type pruneState struct {
	asking bool
	next   model.Plan
	ids    []string
}

func leftOverNotice(count int) string {
	switch count {
	case 0:
		return ""
	case 1:
		return "1 answer is left over from a question that is no longer asked. Press h to review it."
	}
	return fmt.Sprintf("%d answers are left over from questions that are no longer asked. Press h to review them.", count)
}

func (app App) askToPrune() App {
	next, cleared, problems := app.backend.PruneHiddenAnswers(app.plan)
	if hasError(problems) {
		app.status = line{text: plainProblem(problems[0]), tone: toneProblem}
		return app
	}
	if len(cleared.Answers) == 0 {
		app.review.leftOver = 0
		app.status = line{text: nothingLeftOverNotice, tone: toneFaint}
		return app
	}
	app.review.pruning = pruneState{asking: true, next: next, ids: cleared.Answers}
	return app
}

func (app App) updatePruning(key tea.KeyMsg) App {
	switch key.String() {
	case "y":
		return app.confirmPrune()
	case "?":
		app.helpOpen = true
		return app
	}
	app.review.pruning = pruneState{}
	app.status = line{text: "nothing removed", tone: toneFaint}
	return app
}

func (app App) confirmPrune() App {
	state := app.review.pruning
	app.plan = state.next
	keep := ""
	if question, ok := app.current(); ok {
		keep = question.ID
	}
	app = app.refreshed(keep)
	app = app.reviewRebuilt()
	app.status = app.saveStatus(nil)
	if app.status.tone == toneFaint {
		app.status.text = "removed " + pluralWords(len(state.ids), "left over answer", "left over answers") + ". saved"
	}
	return app
}

func (app App) reviewRebuilt() App {
	previous := app.review
	app = app.openReview()
	app.review.focusDocument = previous.focusDocument
	app.review.cursor = max(min(previous.cursor, len(app.review.items)-1), 0)
	app.review = app.review.scrolled(app.documentOffsetLimit(), previous.scroll)
	return app
}

func (app App) leftOverValue(id string) model.Value {
	scope, name, _ := strings.Cut(id, ".")
	if scope == "plan" {
		return app.plan.Answers.Plan[name]
	}
	return app.plan.Answers.Sections[scope][name]
}

func shortValue(value model.Value) string {
	switch value.Kind {
	case model.ValueKindDeferred:
		return "deferred"
	case model.ValueKindFlag:
		if value.Flag {
			return "yes"
		}
		return "no"
	case model.ValueKindList:
		return shortText(strings.Join(value.List, ", "), pluralWords(len(value.List), "item", "items"))
	case model.ValueKindText:
		return shortText(value.Text, "")
	}
	return "no value"
}

func shortText(text string, summary string) string {
	text = singleLine(text)
	if text == "" {
		return "empty"
	}
	runes := []rune(text)
	if len(runes) <= shownValueCharacters {
		return `"` + text + `"`
	}
	hidden := fmt.Sprintf("%d characters", len(runes))
	if summary != "" {
		hidden = summary
	}
	return `"` + string(runes[:hiddenValuePrefix]) + `…" (` + hidden + ")"
}

func (app App) pruningLines(width int) []line {
	ids := app.review.pruning.ids
	content := []line{strongLine("Remove " + pluralWords(len(ids), "left over answer", "left over answers") + "?"), blankLine()}
	content = append(content, lines(toneBody, wrapText("These belong to questions the plan no longer asks, so they are not part of the current plan. They stay in the file until you remove them.", width))...)
	content = append(content, blankLine())
	shown := ids
	if len(ids) > leftOverShownLimit {
		shown = ids[:leftOverShownLimit-1]
	}
	for _, id := range shown {
		content = append(content, lines(toneBody, hanging("  - ", app.questionText(id), width))...)
		content = append(content, lines(toneFaint, hanging("      ", "value: "+shortValue(app.leftOverValue(id)), width))...)
	}
	if len(ids) > len(shown) {
		content = append(content, lines(toneBody, hanging("  - ", fmt.Sprintf("and %d more", len(ids)-len(shown)), width))...)
	}
	return append(content, blankLine(), faintLine("Nothing is removed until you press y."))
}
