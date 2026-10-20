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
	profileANSI      = 2
	profileAscii     = 3
)

var (
	widths  = []int{50, 60, 80, 99, 100, 120, 139, 140, 200}
	heights = []int{10, 24, 50}
)

func everyScreen(t *testing.T) map[string]App {
	t.Helper()
	walking, _ := startWalking(t)
	choosing := moveTo(t, walking, "plan.team_state")
	secret := moveTo(t, walking, "access.repo_url")
	secret = press(secret, typed("https://user:hunter2@git.example.test/r.git"), enter)
	reviewing, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	blocked := press(walking, esc, typed("r"))
	onboardee, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	return map[string]App{
		"walk typing":   walking,
		"walk choosing": choosing,
		"walk problem":  secret,
		"overview":      press(walking, esc),
		"review":        reviewing,
		"review blocks": blocked,
		"review doc":    press(reviewing, tab),
		"sections":      press(choosing, typed("s")),
		"help":          press(choosing, typed("?")),
		"onboardee":     onboardee,
	}
}

func TestNoLineIsWiderThanTheWindow(t *testing.T) {
	t.Parallel()
	for name, app := range everyScreen(t) {
		for _, width := range widths {
			for _, height := range heights {
				sized := resize(app, width, height)
				rows := viewLines(sized)
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
	app, _ := startWalking(t)
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

func TestBothPanesShowAtOneHundredAndFortyColumns(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	view := plainView(resize(app, 140, 40))
	for _, row := range strings.Split(view, "\n")[1:6] {
		if !strings.Contains(row, "│") {
			t.Fatalf("expected two panes, got %q", row)
		}
	}
	if !strings.Contains(view, "Details") || !strings.Contains(view, "Organization") && !strings.Contains(view, "Plan") {
		t.Errorf("the detail and the list should both show:\n%s", view)
	}
	narrow := plainView(resize(app, 99, 40))
	if strings.Contains(narrow, "│") {
		t.Error("99 columns is one column")
	}
	if !strings.Contains(plainView(resize(app, 100, 40)), "│") {
		t.Error("100 columns is two panes")
	}
}

func TestEveryScreenKeepsItsKeyHintsAtEveryWidth(t *testing.T) {
	t.Parallel()
	for name, app := range everyScreen(t) {
		for _, width := range widths[1:] {
			rows := viewLines(resize(app, width, 24))
			if !strings.Contains(rows[len(rows)-1], "help") && !strings.Contains(rows[len(rows)-1], "f1") && !strings.Contains(rows[len(rows)-1], "any key") {
				t.Errorf("%s at %d has no help hint: %q", name, width, rows[len(rows)-1])
			}
		}
	}
}

func TestWithNoColourTheFocusAndKeyHintsAreStillText(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileAscii)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
	for name, app := range everyScreen(t) {
		sized := resize(app, 100, 30)
		if strings.Contains(sized.View(), "\x1b[") {
			t.Fatalf("%s has escape codes with no colour", name)
		}
	}
	screens := everyScreen(t)
	choosing := resize(screens["walk choosing"], 100, 30)
	view := choosing.View()
	for _, want := range []string{"> ", "[ ]", "enter choose", "? help", "* required"} {
		if !strings.Contains(view, want) {
			t.Errorf("walk choosing lacks %q:\n%s", want, view)
		}
	}
	if !strings.Contains(resize(screens["walk problem"], 100, 30).View(), "! problem:") {
		t.Error("a problem needs its text mark")
	}
	if !strings.Contains(resize(screens["review"], 100, 30).View(), "> Items") {
		t.Error("the focused pane needs its mark")
	}
}

func TestWithColourTheAccentIsTheBrandBlue(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileTrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
	app, _ := startWalking(t)
	if !strings.Contains(resize(app, 100, 30).View(), "38;2;0;113;206") {
		t.Error("expected Branded Blue 0071CE in true colour")
	}
	lipgloss.SetColorProfile(profileANSI)
	if strings.Contains(resize(app, 100, 30).View(), "38;2;") {
		t.Error("ANSI terminals get ANSI colours")
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
