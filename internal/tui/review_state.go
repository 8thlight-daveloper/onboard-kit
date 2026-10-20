package tui

import (
	"strings"

	"onboard-kit/internal/model"
)

type reviewKind int

const (
	reviewBlocker reviewKind = iota
	reviewDeferred
	reviewProblem
)

type reviewItem struct {
	kind       reviewKind
	label      string
	text       string
	questionID string
	severity   model.Severity
}

type reviewState struct {
	items         []reviewItem
	cursor        int
	document      []string
	scroll        int
	focusDocument bool
	answered      int
	total         int
}

func newReviewState(backend Backend, plan model.Plan, audience model.Audience, questions []model.Question, document string) reviewState {
	state := reviewState{document: strings.Split(document, "\n"), total: len(questions)}
	for _, question := range questions {
		if stateOf(question.Answer) != "unanswered" {
			state.answered++
		}
	}
	state.items = reviewItems(backend, plan, audience, questions)
	return state
}

func reviewItems(backend Backend, plan model.Plan, audience model.Audience, questions []model.Question) []reviewItem {
	byID := map[string]model.Question{}
	for _, question := range questions {
		byID[question.ID] = question
	}
	var items []reviewItem
	blocked := map[string]bool{}
	for _, blocker := range backend.Blockers(plan, audience) {
		blocked[blocker.Question] = true
		question := byID[blocker.Question]
		label := "required"
		if stateOf(question.Answer) == "deferred" {
			label = "required, deferred"
		}
		items = append(items, reviewItem{kind: reviewBlocker, label: label, text: questionTitle(question, blocker.Question), questionID: blocker.Question})
	}
	for _, question := range questions {
		if stateOf(question.Answer) == "deferred" && !blocked[question.ID] {
			items = append(items, reviewItem{kind: reviewDeferred, label: "deferred", text: questionTitle(question, question.ID), questionID: question.ID})
		}
	}
	for _, problem := range backend.Validate(plan) {
		items = append(items, reviewItem{kind: reviewProblem, label: string(problem.Severity), text: problemText(problem), questionID: problem.Question, severity: problem.Severity})
	}
	return items
}

func questionTitle(question model.Question, fallback string) string {
	if question.Description == "" {
		return fallback
	}
	return question.Description
}

func problemText(problem model.Problem) string {
	if problem.Question == "" {
		return problem.Message
	}
	return problem.Question + ": " + problem.Message
}

func (r reviewState) counts(kind reviewKind) int {
	count := 0
	for _, item := range r.items {
		if item.kind == kind {
			count++
		}
	}
	return count
}

func (r reviewState) moved(delta int) reviewState {
	r.cursor = max(min(r.cursor+delta, len(r.items)-1), 0)
	return r
}

func (r reviewState) scrolled(limit int, delta int) reviewState {
	r.scroll = max(min(r.scroll+delta, limit), 0)
	return r
}

func (r reviewState) selected() (reviewItem, bool) {
	if r.cursor < 0 || r.cursor >= len(r.items) {
		return reviewItem{}, false
	}
	return r.items[r.cursor], true
}

func (a App) documentWidth() int { return layoutFor(a.width, a.height).mainWidth }

func (a App) documentLines() []string {
	var wrapped []string
	for _, source := range a.review.document {
		wrapped = append(wrapped, wrapIndented(source, a.documentWidth())...)
	}
	return wrapped
}

func (a App) documentOffsetLimit() int {
	return max(len(a.documentLines())-layoutFor(a.width, a.height).documentRows, 0)
}
