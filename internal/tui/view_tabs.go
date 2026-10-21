package tui

import (
	"strings"

	"onboard-kit/internal/model"
)

const (
	tabGap      = "  "
	tabGapTight = " "
)

type screenTab struct {
	screen screen
	name   string
	key    string
}

type headerLayout struct {
	keys     bool
	progress bool
	tight    bool
}

func (app App) tabs() []screenTab {
	tabs := []screenTab{{screenWalk, "Walk", "w"}, {screenOverview, "Overview", "o"}, {screenReview, "Review", "r"}}
	if app.canSwitchSections() {
		tabs = append(tabs, screenTab{screenSections, "Sections", "s"})
	}
	if app.canEditDirectory() {
		tabs = append(tabs, screenTab{screenPeople, "People", "P"})
	}
	return tabs
}

func (app App) headerTitle() string {
	if app.audience == model.AudienceOnboardee {
		return "Your onboarding"
	}
	return "Onboarding plan"
}

func (app App) tabStrip(withKeys bool, gap string, colors palette) string {
	var parts []string
	for _, entry := range app.tabs() {
		switch {
		case entry.screen == app.screen && colors.styled():
			parts = append(parts, colors.render(toneHere, true, entry.name))
		case entry.screen == app.screen:
			parts = append(parts, "["+entry.name+"]")
		case withKeys:
			parts = append(parts, colors.render(toneFaint, true, entry.name+" "+entry.key))
		default:
			parts = append(parts, colors.render(toneFaint, true, entry.name))
		}
	}
	return strings.Join(parts, gap)
}

func (app App) headerLeft(layout headerLayout, colors palette) string {
	gap := tabGap
	if layout.tight {
		gap = tabGapTight
	}
	return colors.render(toneStrong, true, app.headerTitle()) + gap + app.tabStrip(layout.keys, gap, colors)
}

func (app App) typingNow() bool {
	return app.screen == screenWalk && app.entry.typing() || app.screen == screenPeople && app.people.typing()
}

func (app App) headerLayout(width int, colors palette) headerLayout {
	candidates := []headerLayout{{true, true, false}, {true, false, false}, {false, true, false}, {false, false, false}, {false, false, true}}
	if app.typingNow() {
		candidates = []headerLayout{{false, true, false}, {false, false, false}, {false, false, true}}
	}
	for _, candidate := range candidates {
		used := cellWidth(app.headerLeft(candidate, colors))
		if candidate.progress {
			used += 2 + cellWidth(app.progressText())
		}
		if used <= width {
			return candidate
		}
	}
	return candidates[len(candidates)-1]
}

func (app App) headerRow(width int, colors palette) string {
	layout := app.headerLayout(width, colors)
	left := app.headerLeft(layout, colors)
	if !layout.progress {
		return left
	}
	right := app.progressText()
	return left + spaces(width-cellWidth(left)-cellWidth(right)) + colors.render(toneFaint, true, right)
}

func (app App) jumpHints(shownInHeader bool) []hint {
	if shownInHeader {
		return nil
	}
	var hints []hint
	for _, entry := range app.tabs() {
		if entry.screen == app.screen || entry.screen == screenWalk && app.screen == screenOverview {
			continue
		}
		hints = append(hints, hint{entry.key, strings.ToLower(entry.name)})
	}
	return hints
}
