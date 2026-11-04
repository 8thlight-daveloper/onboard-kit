package tui

import (
	"strings"

	"onboard-kit/internal/model"
)

const (
	tabGap      = "  "
	tabGapTight = " "
	clipMark    = "…"
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
	clipped  bool
	untitled bool
	windowed bool
}

func (app App) tabs() []screenTab {
	tabs := []screenTab{{screenWalk, "Walk", "w"}, {screenOverview, "Overview", "o"}, {screenReview, "Review", "r"}}
	if app.canSwitchSections() {
		tabs = append(tabs, screenTab{screenSections, "Sections", "s"})
	}
	if app.canEditDirectory() {
		tabs = append(tabs, screenTab{screenContacts, "Contacts", "c"})
	}
	if app.canPrepare() {
		tabs = append(tabs, screenTab{screenPreparation, "Preparation", "t"})
	}
	if app.canInvite() {
		tabs = append(tabs, screenTab{screenInvitation, "Invitation", "i"})
	}
	return tabs
}

func (app App) headerTitle() string {
	if app.audience == model.AudienceOnboardee {
		return "Your onboarding"
	}
	return "Onboarding plan"
}

func (app App) tabParts(withKeys bool, colors palette) ([]string, int) {
	var parts []string
	current := 0
	for index, entry := range app.tabs() {
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
		if entry.screen == app.screen {
			current = index
		}
	}
	return parts, current
}

func (app App) tabStrip(withKeys bool, gap string, colors palette) string {
	parts, _ := app.tabParts(withKeys, colors)
	return strings.Join(parts, gap)
}

func (layout headerLayout) gap() string {
	if layout.tight {
		return tabGapTight
	}
	return tabGap
}

func (app App) headerParts(layout headerLayout, colors palette) ([]string, int) {
	parts, current := app.tabParts(layout.keys, colors)
	if layout.untitled {
		return parts, current
	}
	return append([]string{colors.render(toneStrong, true, app.headerTitle())}, parts...), current + 1
}

func (app App) headerLeft(layout headerLayout, width int, colors palette) string {
	parts, current := app.headerParts(layout, colors)
	if layout.windowed {
		return windowAround(parts, current, layout.gap(), width)
	}
	whole := strings.Join(parts, layout.gap())
	if !layout.clipped || cellWidth(whole) <= width {
		return whole
	}
	shown := parts[0]
	for _, part := range parts[1:] {
		if cellWidth(shown+layout.gap()+part+layout.gap()+clipMark) > width {
			break
		}
		shown += layout.gap() + part
	}
	return shown + layout.gap() + clipMark
}

func windowAround(parts []string, current int, gap string, width int) string {
	first, last := current, current
	shown := func(from int, to int) string {
		text := strings.Join(parts[from:to+1], gap)
		if from > 0 {
			text = clipMark + gap + text
		}
		if to < len(parts)-1 {
			text += gap + clipMark
		}
		return text
	}
	for grown := true; grown; {
		grown = false
		if last < len(parts)-1 && cellWidth(shown(first, last+1)) <= width {
			last++
			grown = true
		}
		if first > 0 && cellWidth(shown(first-1, last)) <= width {
			first--
			grown = true
		}
	}
	return shown(first, last)
}

func (app App) clipKeepsCurrentTab(layout headerLayout, width int, colors palette) bool {
	parts, current := app.headerParts(layout, colors)
	kept := cellWidth(strings.Join(parts[:current+1], layout.gap()))
	if current == len(parts)-1 {
		return kept <= width
	}
	return kept+cellWidth(layout.gap()+clipMark) <= width
}

func (app App) typingNow() bool {
	return app.screen == screenWalk && app.entry.typing() || app.screen == screenContacts && app.contacts.typing() || app.screen == screenPreparation && app.preparationTyping() || app.screen == screenInvitation && app.invitationTyping()
}

func (app App) headerLayout(width int, colors palette) headerLayout {
	candidates := []headerLayout{{keys: true, progress: true}, {keys: true}, {progress: true}, {}, {tight: true}, {tight: true, clipped: true}, {tight: true, untitled: true}, {tight: true, untitled: true, windowed: true}}
	if app.typingNow() {
		candidates = []headerLayout{{progress: true}, {}, {tight: true}, {tight: true, clipped: true}, {tight: true, untitled: true}, {tight: true, untitled: true, windowed: true}}
	}
	for _, candidate := range candidates {
		if candidate.clipped {
			if app.clipKeepsCurrentTab(candidate, width, colors) {
				return candidate
			}
			continue
		}
		used := cellWidth(app.headerLeft(candidate, width, colors))
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
	left := app.headerLeft(layout, width, colors)
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
