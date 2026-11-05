package tui

import (
	"os"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/decoder"
)

func acmePlan(t *testing.T) model.Plan {
	t.Helper()
	data, err := os.ReadFile("../../docs/reference/examples/acme/acme.obk.yaml")
	if err != nil {
		t.Fatal(err)
	}
	plan, problems := decoder.New().Decode(data)
	if len(problems) > 0 {
		t.Fatalf("the ACME plan has problems: %v", problems)
	}
	return plan
}

func blankPlan(t *testing.T) model.Plan {
	t.Helper()
	plan := acmePlan(t)
	plan.Answers = model.Answers{}
	return plan
}

type recorder struct {
	saved []model.Plan
}

func (recording *recorder) save(plan model.Plan) error {
	recording.saved = append(recording.saved, plan)
	return nil
}

func startWith(t *testing.T, plan model.Plan, audience model.Audience) (App, *recorder) {
	t.Helper()
	recorded := &recorder{}
	app := NewApp(NewBackend(), plan, recorded.save, audience)
	app.rail = unicodeRailGlyphs()
	return app, recorded
}

func startAnswering(t *testing.T) (App, *recorder) {
	t.Helper()
	return startWith(t, blankPlan(t), model.AudienceOnboarder)
}

func press(app App, keys ...tea.KeyMsg) App {
	for _, key := range keys {
		next, _ := app.Update(key)
		app = next.(App)
	}
	return app
}

func resize(app App, width int, height int) App {
	next, _ := app.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return next.(App)
}

func typed(text string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)}
}

var (
	enter    = tea.KeyMsg{Type: tea.KeyEnter}
	down     = tea.KeyMsg{Type: tea.KeyDown}
	up       = tea.KeyMsg{Type: tea.KeyUp}
	space    = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")}
	esc      = tea.KeyMsg{Type: tea.KeyEsc}
	shiftTab = tea.KeyMsg{Type: tea.KeyShiftTab}
	tab      = tea.KeyMsg{Type: tea.KeyTab}
	ctrlP    = tea.KeyMsg{Type: tea.KeyCtrlP}
	ctrlD    = tea.KeyMsg{Type: tea.KeyCtrlD}
	left     = tea.KeyMsg{Type: tea.KeyLeft}
	right    = tea.KeyMsg{Type: tea.KeyRight}
)

func moveTo(t *testing.T, app App, id string) App {
	t.Helper()
	index := app.indexOf(id)
	if index < 0 {
		t.Fatalf("question %s is not asked", id)
	}
	return app.openQuestion(index)
}

var escapeCodes = regexp.MustCompile("\x1b\\[[0-9;?]*[a-zA-Z]|\x1b\\][^\x07\x1b]*(\x07|\x1b\\\\)")

func plainView(app App) string {
	return escapeCodes.ReplaceAllString(app.View(), "")
}

func viewLines(app App) []string {
	return strings.Split(plainView(app), "\n")
}

var sgrSequence = regexp.MustCompile("\x1b\\[([0-9;]*)m")

func sgrBefore(t *testing.T, view string, text string) string {
	t.Helper()
	index := strings.Index(view, text)
	if index < 0 {
		t.Fatalf("%q is not on the screen:\n%s", text, plainText(view))
	}
	found := sgrSequence.FindAllStringSubmatch(view[:index], -1)
	if len(found) == 0 || found[len(found)-1][1] == "0" || found[len(found)-1][1] == "" {
		return ""
	}
	return found[len(found)-1][1]
}

func isFaint(t *testing.T, view string, text string) bool {
	t.Helper()
	for _, parameter := range strings.Split(sgrBefore(t, view, text), ";") {
		if parameter == "2" {
			return true
		}
	}
	return false
}

func plainText(view string) string { return escapeCodes.ReplaceAllString(view, "") }
