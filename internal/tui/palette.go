package tui

import "github.com/charmbracelet/lipgloss"

type tone int

const (
	tonePlain tone = iota
	toneHeading
	toneFocus
	toneMuted
	toneWarning
	toneProblem
)

type palette struct {
	styles map[tone]lipgloss.Style
	key    lipgloss.Style
}

func accentColor() lipgloss.TerminalColor {
	return lipgloss.CompleteColor{TrueColor: "#0071CE", ANSI256: "32", ANSI: "4"}
}

func newPalette() palette {
	warning := lipgloss.CompleteColor{TrueColor: "#FF8300", ANSI256: "208", ANSI: "3"}
	problem := lipgloss.CompleteColor{TrueColor: "#F32735", ANSI256: "196", ANSI: "1"}
	return palette{
		styles: map[tone]lipgloss.Style{
			tonePlain:   lipgloss.NewStyle(),
			toneHeading: lipgloss.NewStyle().Bold(true).Foreground(accentColor()),
			toneFocus:   lipgloss.NewStyle().Bold(true).Foreground(accentColor()),
			toneMuted:   lipgloss.NewStyle().Faint(true),
			toneWarning: lipgloss.NewStyle().Foreground(warning),
			toneProblem: lipgloss.NewStyle().Foreground(problem),
		},
		key: lipgloss.NewStyle().Bold(true).Foreground(accentColor()),
	}
}

func (p palette) style(of tone) lipgloss.Style { return p.styles[of] }
