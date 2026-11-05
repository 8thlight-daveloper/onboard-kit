package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"onboard-kit/internal/model"
)

const (
	profileTrueColor = 0
	profileANSI256   = 1
	profileANSI      = 2
	profileAscii     = 3
)

var (
	widths  = []int{60, 70, 80, 99, 100, 119, 120, 139, 140, 160, 200}
	heights = []int{12, 13, 16, 24, 30, 40, 50}
)

func everyScreen(t *testing.T) map[string]App {
	t.Helper()
	answering, _ := startAnswering(t)
	choosing := moveTo(t, answering, "plan.team_state")
	secret := moveTo(t, answering, "access.repo_url")
	secret = press(secret, typed("https://user:hunter2@git.example.test/r.git"), enter)
	reviewing, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	blocked := press(answering, esc, typed("r"))
	onboardee, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	owner := moveTo(t, answering, "access.onboarding_owner_role")
	contacts := moveTo(t, answering, "stuck-general.general_contacts")
	emptyPlan, _ := startWith(t, withoutDirectory(blankPlan(t)), model.AudienceOnboarder)
	emptyOnboardee, _ := startWith(t, withoutDirectory(blankPlan(t)), model.AudienceOnboardee)
	subsectionFake, subsectionPlan := subsectionBackend()
	subsections := startFakeIn(subsectionFake, subsectionPlan, model.AudienceOnboarder).openOverview()
	flagged := press(reviewing.openOverview(), typed("c"))
	emptyContacts := press(emptyPlan.openOverview(), typed("c"))
	screens := map[string]App{
		"contacts list":                flagged,
		"contacts roles":               press(flagged, right),
		"contacts teams":               press(flagged, right, right),
		"contacts form":                press(flagged, typed("a"), typed("Pat Lee")),
		"contacts problem":             press(flagged, typed("a"), ctrlS),
		"contacts holders":             press(flagged, right, typed("a"), typed("Lead"), enter, down, space),
		"contacts rows":                press(flagged, right, right, typed("a"), typed("Team"), enter, down),
		"contacts override":            press(flagged, right, right, typed("a"), typed("Team"), enter, down, enter, down, space),
		"contacts team detail":         press(flagged, right, right),
		"contacts team contact":        press(flagged, right, right, typed("a"), typed("Team"), enter),
		"contacts team contact picker": press(flagged, right, right, typed("a"), typed("Team"), enter, enter),
		"contacts remove team contact": press(flagged, down, down, down, typed("x")),
		"contacts remove":              press(flagged, down, typed("x")),
		"contacts empty":               emptyContacts,
		"contacts help":                press(flagged, typed("?")),
		"answer typing":                answering,
		"answer blurred":               press(answering, typed("Glob"), esc),
		"answer options":               press(answering, down, down, down),
		"answer strip":                 press(answering, up, right),
		"answer strip help":            press(answering, up, typed("?")),
		"answer choosing":              choosing,
		"answer problem":               secret,
		"overview":                     press(answering, esc, esc),
		"review":                       reviewing,
		"review blocks":                blocked,
		"review doc":                   press(reviewing, right),
		"review left over":             startReviewingFrom(t, planWithLeftOvers(t)),
		"review prune":                 press(startReviewingFrom(t, planWithLeftOvers(t)), typed("h")),
		"review prune help":            press(startReviewingFrom(t, planWithLeftOvers(t)), typed("h"), typed("?")),
		"sections":                     press(choosing, typed("s")),
		"help":                         press(choosing, typed("?")),
		"onboardee":                    onboardee,

		"pick one":             owner,
		"pick many":            press(contacts, typed("j"), space, typed("j"), space),
		"pick empty":           moveTo(t, emptyPlan, "access.onboarding_owner_role"),
		"pick empty onboardee": emptyOnboardee,
		"pick help":            press(contacts, typed("?")),
		"subsections":          subsections,
		"sections with tabs":   press(owner, typed("s")),
	}
	for name, app := range longScreens() {
		screens[name] = app
	}
	for name, steps := range preparationSteps() {
		screens[strings.ReplaceAll(name, "-", " ")] = press(onPreparation(t, withSamLee(t)), steps...)
	}
	for name, steps := range invitationSteps() {
		screens[strings.ReplaceAll(name, "-", " ")] = press(onInvitation(t, newFakeInvitations(t)), steps...)
	}
	return screens
}

func TestNoLineIsWiderThanTheWindow(t *testing.T) {
	t.Parallel()
	for name, app := range everyScreen(t) {
		for _, width := range widths {
			for _, height := range heights {
				rows := viewLines(resize(app, width, height))
				if len(rows) > height {
					t.Errorf("%s at %dx%d draws %d rows", name, width, height, len(rows))
				}
				for index, row := range rows {
					if lipgloss.Width(row) > width {
						t.Errorf("%s at %dx%d row %d is %d wide: %q", name, width, height, index, lipgloss.Width(row), row)
					}
				}
			}
		}
	}
}

func TestAWindowThatIsTooSmallGetsAPlainMessage(t *testing.T) {
	t.Parallel()
	app, _ := startAnswering(t)
	for _, size := range [][2]int{{50, 24}, {59, 40}, {80, 11}, {200, 10}} {
		view := plainView(resize(app, size[0], size[1]))
		if !strings.Contains(strings.Join(strings.Fields(view), " "), "larger") {
			t.Errorf("%dx%d should ask for a larger window, got %q", size[0], size[1], view)
		}
	}
	if strings.Contains(plainView(resize(app, 60, 12)), "larger") {
		t.Error("60x12 is large enough")
	}
}

func TestTheAnswerScreenHasTwoColumnsFromOneHundredAndTwentyAndOneBelow(t *testing.T) {
	t.Parallel()
	app, _ := startAnswering(t)
	for _, width := range []int{120, 140, 200} {
		view := plainView(resize(app, width, 40))
		if !strings.Contains(view, "│") || !strings.Contains(view, "About row 1") {
			t.Errorf("%d columns should show the answer and the notes side by side:\n%s", width, view)
		}
	}
	for _, width := range []int{80, 99, 119} {
		view := plainView(resize(app, width, 40))
		if strings.Contains(view, "│") {
			t.Errorf("%d columns is one column:\n%s", width, view)
		}
		if !strings.Contains(view, "About row 1") {
			t.Errorf("%d columns should still show the notes below the answer", width)
		}
	}
}

func TestTheQuestionListIsNotOnTheAnswerScreen(t *testing.T) {
	t.Parallel()
	app, _ := startAnswering(t)
	view := plainView(resize(app, 140, 40))
	for _, title := range []string{"Where to request source control access", "Who can help with anything"} {
		if strings.Contains(view, title) {
			t.Errorf("the Answer screen should show one question, found %q", title)
		}
	}
}

func TestEveryScreenKeepsItsKeyHintsAtEveryWidth(t *testing.T) {
	t.Parallel()
	for name, app := range everyScreen(t) {
		for _, width := range widths {
			rows := viewLines(resize(app, width, 24))
			last := rows[len(rows)-1]
			if !strings.Contains(last, "help") && !strings.Contains(last, "f1") && !strings.Contains(last, "any key") {
				t.Errorf("%s at %d has no help hint: %q", name, width, last)
			}
		}
	}
}

func TestWithNoColourTheMarksAndKeyHintsAreStillText(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileAscii)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
	for name, app := range everyScreen(t) {
		if strings.Contains(resize(app, 120, 30).View(), "\x1b[") {
			t.Fatalf("%s has escape codes with no colour", name)
		}
	}
	screens := everyScreen(t)
	view := resize(screens["answer choosing"], 120, 30).View()
	for _, want := range []string{"> established", "◆", "enter choose", "? help", "required", "not answered yet", "2 of 8"} {
		if !strings.Contains(view, want) {
			t.Errorf("answer choosing lacks %q:\n%s", want, view)
		}
	}
	if !strings.Contains(resize(screens["answer problem"], 120, 30).View(), "! problem:") {
		t.Error("a problem needs its text mark")
	}
	if !strings.Contains(resize(screens["review"], 120, 30).View(), "> Items") {
		t.Error("the focused pane needs its mark")
	}
	if !strings.Contains(resize(screens["overview"], 120, 30).View(), "> [ ] *") {
		t.Error("the overview cursor needs its mark")
	}
}

func TestNoScreenUsesAFixedColourAndTheAccentMarksPlaceNotTitles(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileTrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
	app, _ := startAnswering(t)
	contacts, _ := onContacts(t, contactsWithFlags(t))
	screens := map[string]App{
		"answer":   moveTo(t, app, "plan.team_state"),
		"overview": press(app, esc, esc),
		"review":   press(app, esc, typed("r")),
		"contacts": contacts,
	}
	for name, shown := range screens {
		view := resize(shown, 120, 30).View()
		for _, fixed := range []string{"38;2;", "48;2;", "38;5;", "48;5;"} {
			if strings.Contains(view, fixed) {
				t.Errorf("%s: the screen uses a fixed colour sequence %q", name, fixed)
			}
		}
	}
	view := resize(screens["answer"], 120, 30).View()
	if !strings.Contains(view, "32m") {
		t.Error("expected the terminal green accent")
	}
	for _, title := range []string{"Building an onboarding plan", "Plan", "Is the team established"} {
		if strings.Contains(sgrBefore(t, view, title), "32") {
			t.Errorf("the accent must not colour the text %q", title)
		}
	}
}

func TestTextIsWrappedAtThePaneWidth(t *testing.T) {
	t.Parallel()
	text := "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda " + strings.Repeat("x", 30)
	for _, width := range []int{1, 5, 12, 20, 40} {
		wrapped := wrapText(text, width)
		for _, row := range wrapped {
			if lipgloss.Width(row) > width {
				t.Errorf("width %d: %q is %d wide", width, row, lipgloss.Width(row))
			}
		}
		if strings.Join(strings.Fields(strings.Join(wrapped, " ")), "") != strings.Join(strings.Fields(text), "") {
			t.Errorf("width %d lost text: %q", width, wrapped)
		}
	}
	if got := wrapText("a\n\nb", 10); fmt.Sprint(got) != "[a  b]" {
		t.Errorf("paragraphs should stay apart: %q", got)
	}
}

func TestLongAnswersAndDescriptionsAreCutNotOverflowed(t *testing.T) {
	t.Parallel()
	if got := truncateCells("abcdefghij", 5); got != "abcd…" {
		t.Fatalf("got %q", got)
	}
	if got := truncateCells("abc", 5); got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestHintsDropTheLeastUsefulFirstAndKeepHelp(t *testing.T) {
	t.Parallel()
	hints := []hint{{"enter", "choose"}, {"p", "back"}, {"q", "quit"}, {"?", "help"}}
	fitted := fitHints(hints, 14)
	if got := hintText(fitted); got != "enter choose  ? help" && got != "? help" {
		t.Fatalf("got %q", got)
	}
	if last := fitted[len(fitted)-1]; last.key != "?" {
		t.Fatal("help must stay")
	}
}
