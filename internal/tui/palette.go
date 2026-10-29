package tui

import "github.com/charmbracelet/lipgloss"

type tone int

const (
	toneBody tone = iota
	toneStrong
	toneHere
	toneFaint
	toneWarning
	toneProblem
)

type palette struct {
	styles map[tone]lipgloss.Style
}

func terminalGreen() lipgloss.TerminalColor {
	return lipgloss.ANSIColor(2)
}

func terminalYellow() lipgloss.TerminalColor {
	return lipgloss.ANSIColor(3)
}

func terminalRed() lipgloss.TerminalColor {
	return lipgloss.ANSIColor(1)
}

func newPalette() palette {
	return palette{styles: map[tone]lipgloss.Style{
		toneBody:    lipgloss.NewStyle(),
		toneStrong:  lipgloss.NewStyle().Bold(true),
		toneHere:    lipgloss.NewStyle().Bold(true).Foreground(terminalGreen()),
		toneFaint:   lipgloss.NewStyle().Faint(true),
		toneWarning: lipgloss.NewStyle().Foreground(terminalYellow()),
		toneProblem: lipgloss.NewStyle().Foreground(terminalRed()),
	}}
}

func (colors palette) render(weight tone, focused bool, text string) string {
	style := colors.styles[weight]
	if !focused {
		style = style.Faint(true)
	}
	return style.Render(text)
}

func (colors palette) styled() bool {
	return colors.styles[toneStrong].Render("x") != "x"
}
