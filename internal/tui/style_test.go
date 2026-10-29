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

func withoutColour(t *testing.T) {
	t.Helper()
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileAscii)
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
		expectFaint(t, view, "2 of 8")
	}
}

func TestOnTheOverviewTheListIsBrightAndThePreviewIsFaint(t *testing.T) {
	withColour(t)
	app, _ := startWalking(t)
	view := resize(press(app, esc), 140, 40).View()
	expectBright(t, view, "Where do people request source control access?")
	expectBright(t, view, "What is the organization's name?")
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
		for _, title := range []string{"Onboarding plan", "Blockers"} {
			if !strings.Contains(view, title) {
				continue
			}
			if sgrBefore(t, view, title) == "32" || sgrBefore(t, view, title) == "1;32" {
				t.Errorf("%s: %q is drawn in the accent", name, title)
			}
		}
	}
}

func TestOnTheContactsScreenTheListIsBrightAndTheDetailIsFaint(t *testing.T) {
	withColour(t)
	app, _ := onContacts(t, contactsWithFlags(t))
	view := resize(app, 140, 40).View()
	expectBright(t, view, "Alex Rivera")
	expectBright(t, view, "Contacts 6")
	expectFaint(t, view, "Roles 4")
	expectFaint(t, view, "About this contact")
	expectFaint(t, view, "Id: alex-rivera")
	if got := sgrBefore(t, view, "Contacts 6"); got != "1;32" {
		t.Errorf("the current list is bold in the accent, got SGR %q", got)
	}
	flagged := resize(press(app, typed("G")), 140, 40).View()
	if got := sgrBefore(t, flagged, "warning:"); got != "33" {
		t.Errorf("a warning is the terminal yellow, got SGR %q", got)
	}
	unselected := resize(press(app, typed("g")), 140, 40).View()
	if got := sgrBefore(t, unselected, "! no way to reach"); got != "33" {
		t.Errorf("a flagged row is the terminal yellow, got SGR %q", got)
	}
}

func TestWithNoColourTheContactsScreenKeepsItsMarks(t *testing.T) {
	withoutColour(t)
	app, _ := onContacts(t, contactsWithFlags(t))
	view := resize(app, 120, 30).View()
	for _, want := range []string{"[Contacts 6]", "> Alex Rivera", "! email looks wrong", "! no way to reach", "[Contacts]", "a add", "x remove"} {
		if !strings.Contains(view, want) {
			t.Errorf("lacks %q:\n%s", want, view)
		}
	}
	if !strings.Contains(resize(press(app, tab), 120, 30).View(), "! vacant") {
		t.Error("a vacant role keeps its text mark")
	}
	problem := resize(press(app, typed("a"), ctrlS), 120, 30).View()
	if !strings.Contains(problem, "! problem: A name is needed.") || !strings.Contains(problem, "> Name") {
		t.Errorf("a form keeps its focus mark and problem mark:\n%s", problem)
	}
}
