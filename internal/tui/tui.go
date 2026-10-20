package tui

import (
	"io"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

func Run(backend Backend, plan model.Plan, save func(model.Plan) error, audience model.Audience, input io.Reader, output io.Writer) error {
	program := tea.NewProgram(NewApp(backend, plan, save, audience), tea.WithInput(input), tea.WithOutput(output), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
