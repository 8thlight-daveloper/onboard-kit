package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

func preparationSteps() map[string][]tea.KeyMsg {
	staged := []tea.KeyMsg{enter, enter, typed(" Jr"), ctrlS}
	return map[string][]tea.KeyMsg{
		"preparation-people":      nil,
		"preparation-add":         {typed("a"), typed("alex-doe")},
		"preparation-person":      {enter},
		"preparation-recipient":   {enter, enter, typed(" Jr")},
		"preparation-pick":        {enter, enter, down, down, down, down, enter},
		"preparation-arrangement": {enter, typed("a"), typed("second-pairing"), down},
		"preparation-helpers":     {enter, down, down, enter, down, down, down, enter, typed("j"), space},
		"preparation-problem":     {enter, down, enter, down, down, enter, up, enter, down, down, down, down, enter, down, enter, ctrlS},
		"preparation-remove":      {enter, down, typed("x")},
		"preparation-drop":        append(append([]tea.KeyMsg{}, staged...), esc),
		"preparation-review":      append(append([]tea.KeyMsg{}, staged...), typed("G"), enter),
		"preparation-guide":       append(append([]tea.KeyMsg{}, staged...), typed("G"), enter, tab),
	}
}

func preparing(t *testing.T, name string) func(goldenSize) App {
	t.Helper()
	return func(size goldenSize) App {
		app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
		app = resize(app.WithPreparation(withSamLee(t)), size.width, size.height)
		return press(app, append([]tea.KeyMsg{typed("t")}, preparationSteps()[name]...)...)
	}
}
