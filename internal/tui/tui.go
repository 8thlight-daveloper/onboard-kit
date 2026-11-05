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

func Run(backend Backend, plan model.Plan, save func(model.Plan) error, audience model.Audience, notice string, preparation PreparationPort, invitation InvitationPort, about *About, input io.Reader, output io.Writer) error {
	app := NewApp(backend, plan, save, audience).WithNotice(notice).WithPreparation(preparation).WithInvitation(invitation).WithAbout(about)
	_, err := runProgram(app, input, output)
	return err
}

func RunFrontDoor(door FrontDoor, input io.Reader, output io.Writer) (FrontDoorChoice, error) {
	final, err := runProgram(door, input, output)
	if err != nil {
		return FrontDoorChoice{}, err
	}
	return final.(FrontDoor).Choice(), nil
}

func runProgram(shown tea.Model, input io.Reader, output io.Writer) (tea.Model, error) {
	terminations := make(chan os.Signal, 1)
	signal.Notify(terminations, syscall.SIGTERM)
	defer signal.Stop(terminations)
	program := tea.NewProgram(shown, tea.WithInput(input), tea.WithOutput(output), tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		return final, err
	}
	select {
	case <-terminations:
		return final, ErrTerminated
	default:
		return final, nil
	}
}
