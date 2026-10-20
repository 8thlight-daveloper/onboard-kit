package tui

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"onboard-kit/internal/model"
)

var update = flag.Bool("update", false, "rewrite the screen files under testdata/screens")

type goldenSize struct {
	width  int
	height int
}

var goldenSizes = []goldenSize{{80, 24}, {140, 40}}

func goldenScreens(t *testing.T) map[string]func(size goldenSize) App {
	t.Helper()
	build := func(plan model.Plan, audience model.Audience, steps ...func(App) App) func(goldenSize) App {
		return func(size goldenSize) App {
			app, _ := startWith(t, plan, audience)
			app = resize(app, size.width, size.height)
			for _, step := range steps {
				app = step(app)
			}
			return app
		}
	}
	keys := func(pressed ...tea.KeyMsg) func(App) App {
		return func(app App) App { return press(app, pressed...) }
	}
	goTo := func(id string) func(App) App {
		return func(app App) App { return moveTo(t, app, id) }
	}
	return map[string]func(goldenSize) App{
		"walk":      build(blankPlan(t), model.AudienceOnboarder, goTo("plan.team_state")),
		"edit":      build(blankPlan(t), model.AudienceOnboarder, keys(typed("Glob"))),
		"overview":  build(acmePlan(t), model.AudienceOnboarder, goTo("plan.team_state"), keys(typed("o"))),
		"review":    build(acmePlan(t), model.AudienceOnboarder, goTo("plan.team_state"), keys(typed("d"), typed("r"))),
		"sections":  build(blankPlan(t), model.AudienceOnboarder, goTo("plan.team_state"), keys(typed("s"))),
		"onboardee": build(blankPlan(t), model.AudienceOnboardee),
		"too-small": func(size goldenSize) App {
			app, _ := startWith(t, blankPlan(t), model.AudienceOnboarder)
			return resize(app, size.width, 10)
		},
	}
}

func TestScreensMatchTheirWireframes(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileAscii)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
	for name, build := range goldenScreens(t) {
		for _, size := range goldenSizes {
			path := filepath.Join("testdata", "screens", fmt.Sprintf("%s.%d.txt", name, size.width))
			got := wireframe(build(size))
			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				continue
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%s is missing, run go test ./internal/tui -run Wireframes -update: %v", path, err)
			}
			if string(want) != got {
				t.Errorf("%s differs from the screen.\nwant:\n%s\ngot:\n%s", path, want, got)
			}
		}
	}
}

func wireframe(app App) string {
	rows := viewLines(app)
	for index, row := range rows {
		rows[index] = strings.TrimRight(row, " ")
	}
	return strings.Join(rows, "\n") + "\n"
}
