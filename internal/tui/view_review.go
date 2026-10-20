package tui

import "fmt"

func (a App) reviewBody(l layout, p palette) []string {
	items, focus := a.reviewItemLines(l.listWidthFor(), a.review.cursor)
	items = append([]line{a.paneTitle("Items", !a.review.focusDocument)}, scrollWindow(items, focus, l.documentRows)...)
	document := append([]line{a.paneTitle(a.documentTitle(), a.review.focusDocument)}, a.documentView(l)...)
	switch {
	case l.twoPanes:
		return sideBySide(items, document, l, p)
	case a.review.focusDocument:
		return singlePane(document, l, p)
	}
	return singlePane(items, l, p)
}

func (l layout) listWidthFor() int {
	if l.twoPanes {
		return l.listWidth
	}
	return l.width
}

func (a App) paneTitle(name string, focused bool) line {
	if focused {
		return line{text: "> " + name, tone: toneFocus}
	}
	return line{text: "  " + name, tone: toneMuted}
}

func (a App) documentTitle() string {
	limit := a.documentOffsetLimit()
	if limit == 0 {
		return "Document, all shown"
	}
	return fmt.Sprintf("Document, %d%%", a.review.scroll*100/limit)
}

func (a App) documentView(l layout) []line {
	wrapped := a.documentLines()
	first := min(a.review.scroll, len(wrapped))
	last := min(first+l.documentRows, len(wrapped))
	return lines(tonePlain, wrapped[first:last])
}

func (a App) reviewItemLines(width int, selected int) ([]line, int) {
	r := a.review
	content := []line{plainLine(fmt.Sprintf("Answered %d of %d.", r.answered, r.total))}
	if len(r.items) == 0 {
		content = append(content, plainLine("Nothing is waiting."))
		return content, 0
	}
	focus := 0
	groups := []struct {
		kind  reviewKind
		title string
	}{{reviewBlocker, "Blockers"}, {reviewDeferred, "Deferred"}, {reviewProblem, "Problems"}}
	for _, group := range groups {
		if r.counts(group.kind) == 0 {
			continue
		}
		content = append(content, plainLine(""), headingLine(fmt.Sprintf("%s (%d)", group.title, r.counts(group.kind))))
		for index, item := range r.items {
			if item.kind != group.kind {
				continue
			}
			if index == selected {
				focus = len(content)
			}
			content = append(content, a.itemLines(item, index == selected, width)...)
		}
	}
	return content, focus
}

func (a App) itemLines(item reviewItem, selected bool, width int) []line {
	tone := tonePlain
	switch {
	case selected:
		tone = toneFocus
	case item.severity == "error":
		tone = toneProblem
	case item.severity == "warning":
		tone = toneWarning
	}
	marker := "  "
	if selected {
		marker = "> "
	}
	rows := hanging(marker, item.label+": "+item.text, width)
	return lines(tone, rows)
}
