package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

func invitationSteps() map[string][]tea.KeyMsg {
	chosen := []tea.KeyMsg{down, enter, enter, enter}
	return map[string][]tea.KeyMsg{
		"invitation-people":  nil,
		"invitation-blocked": {enter},
		"invitation-form":    {down, enter},
		"invitation-pick":    {down, enter, enter},
		"invitation-helpers": append(append([]tea.KeyMsg{}, chosen...), down, enter, down, space, down, down, down, space),
		"invitation-missing": {down, enter, ctrlS},
		"invitation-confirm": append(append([]tea.KeyMsg{}, chosen...), ctrlS),
		"invitation-written": append(append([]tea.KeyMsg{}, chosen...), ctrlS, typed("y")),
		"invitation-help":    append(append([]tea.KeyMsg{}, chosen...), ctrlS, typed("?")),
	}
}

func inviting(t *testing.T, name string) func(goldenSize) App {
	t.Helper()
	return func(size goldenSize) App {
		app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
		app = resize(app.WithPreparation(withSamLee(t)).WithInvitation(newFakeInvitations(t)), size.width, size.height)
		return press(app, append([]tea.KeyMsg{typed("i")}, invitationSteps()[name]...)...)
	}
}
