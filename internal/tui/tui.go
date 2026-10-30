package tui

import (
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

var (
	ErrInterrupted = tea.ErrInterrupted
	ErrTerminated  = errors.New("program was terminated")
)

func Run(backend Backend, plan model.Plan, save func(model.Plan) error, audience model.Audience, notice string, input io.Reader, output io.Writer) error {
	terminations := make(chan os.Signal, 1)
	signal.Notify(terminations, syscall.SIGTERM)
	defer signal.Stop(terminations)
	program := tea.NewProgram(NewApp(backend, plan, save, audience).WithNotice(notice), tea.WithInput(input), tea.WithOutput(output), tea.WithAltScreen())
	_, err := program.Run()
	if err != nil {
		return err
	}
	select {
	case <-terminations:
		return ErrTerminated
	default:
		return nil
	}
}
