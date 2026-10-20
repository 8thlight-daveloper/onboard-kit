package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"onboard-kit/internal/model"
)

func withColour(t *testing.T) {
	t.Helper()
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
}

func expectBright(t *testing.T, view string, text string) {
	t.Helper()
	if isFaint(t, view, text) {
		t.Errorf("%q should be drawn bright", text)
	}
}

func expectFaint(t *testing.T, view string, text string) {
	t.Helper()
	if !isFaint(t, view, text) {
		t.Errorf("%q should be drawn faint", text)
	}
}

func TestOnTheWalkTheAnswerIsBrightAndTheNotesAreFaint(t *testing.T) {
	withColour(t)
	app, _ := startWalking(t)
	for _, width := range []int{80, 120, 140} {
		view := resize(moveTo(t, app, "plan.team_state"), width, 40).View()
		expectBright(t, view, "Is the team established")
		expectBright(t, view, "deferred, answer later")
		expectFaint(t, view, "Type: choice")
		expectFaint(t, view, "Answered by")
		expectFaint(t, view, "2 of 9")
	}
}

func TestOnTheOverviewTheListIsBrightAndThePreviewIsFaint(t *testing.T) {
	withColour(t)
	app, _ := startWalking(t)
	view := resize(press(app, esc), 140, 40).View()
	expectBright(t, view, "Where to request source control access")
	expectBright(t, view, "Name of the organization")
	expectFaint(t, view, "About this question")
	expectFaint(t, view, "Type: string")
}

func TestOnTheReviewTheFocusedPaneIsBrightAndTheOtherIsFaint(t *testing.T) {
	withColour(t)
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	items := resize(app, 140, 40).View()
	expectBright(t, items, "Answered ")
	expectBright(t, items, "Items")
	expectFaint(t, items, "Document, ")
	document := resize(press(app, tab), 140, 40).View()
	expectFaint(t, document, "Answered ")
	expectFaint(t, document, "Items")
	expectBright(t, document, "Document, ")
}

func TestTheAccentNeverColoursATitle(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileANSI)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
	app, _ := startWalking(t)
	for name, shown := range map[string]App{"walk": app, "overview": press(app, esc), "review": press(app, esc, typed("r"))} {
		view := resize(shown, 140, 40).View()
		for _, title := range []string{"Onboarding plan", "Blockers", "Overview"} {
			if !strings.Contains(view, title) {
				continue
			}
			if sgrBefore(t, view, title) == "34" || sgrBefore(t, view, title) == "1;34" {
				t.Errorf("%s: %q is drawn in the accent", name, title)
			}
		}
	}
}
