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

func brandBlue() lipgloss.TerminalColor {
	return lipgloss.CompleteColor{TrueColor: "#0071CE", ANSI256: "32", ANSI: "4"}
}

func brandOrange() lipgloss.TerminalColor {
	return lipgloss.CompleteColor{TrueColor: "#FF8300", ANSI256: "208", ANSI: "3"}
}

func brandRed() lipgloss.TerminalColor {
	return lipgloss.CompleteColor{TrueColor: "#F32735", ANSI256: "196", ANSI: "1"}
}

func newPalette() palette {
	return palette{styles: map[tone]lipgloss.Style{
		toneBody:    lipgloss.NewStyle(),
		toneStrong:  lipgloss.NewStyle().Bold(true),
		toneHere:    lipgloss.NewStyle().Bold(true).Foreground(brandBlue()),
		toneFaint:   lipgloss.NewStyle().Faint(true),
		toneWarning: lipgloss.NewStyle().Foreground(brandOrange()),
		toneProblem: lipgloss.NewStyle().Foreground(brandRed()),
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
