package tui

import "strings"

type line struct {
	text string
	tone tone
	raw  bool
}

type pane struct {
	lines   []line
	focused bool
}

func bodyLine(text string) line   { return line{text: text} }
func faintLine(text string) line  { return line{text: text, tone: toneFaint} }
func strongLine(text string) line { return line{text: text, tone: toneStrong} }
func blankLine() line             { return line{} }

func lines(weight tone, texts []string) []line {
	result := make([]line, 0, len(texts))
	for _, text := range texts {
		result = append(result, line{text: text, tone: weight})
	}
	return result
}

func renderCell(entry line, width int, focused bool, colors palette) string {
	if entry.raw {
		return entry.text + spaces(width-cellWidth(entry.text))
	}
	text := truncateCells(singleLine(entry.text), width)
	return colors.render(entry.tone, focused, text) + spaces(width-cellWidth(text))
}

func (shown pane) cell(index int, width int, colors palette) string {
	if index >= len(shown.lines) {
		return spaces(width)
	}
	return renderCell(shown.lines[index], width, shown.focused, colors)
}

func stackPane(shown pane, width int, height int, colors palette) []string {
	rows := make([]string, 0, height)
	for index := 0; index < height; index++ {
		rows = append(rows, strings.TrimRight(shown.cell(index, width, colors), " "))
	}
	return rows
}

func joinPanes(left pane, right pane, leftWidth int, rightWidth int, height int, colors palette) []string {
	divider := " " + colors.render(toneFaint, true, "│") + " "
	filled := max(len(left.lines), len(right.lines))
	rows := make([]string, 0, height)
	for index := 0; index < height; index++ {
		separator := "   "
		if index < filled {
			separator = divider
		}
		row := left.cell(index, leftWidth, colors) + separator + right.cell(index, rightWidth, colors)
		rows = append(rows, strings.TrimRight(row, " "))
	}
	return rows
}

func scrollWindow(content []line, focus int, height int) []line {
	if len(content) <= height {
		return content
	}
	first := max(min(focus-height/2, len(content)-height), 0)
	return content[first : first+height]
}

func clipLines(content []line, height int) []line {
	if len(content) <= height {
		return content
	}
	return content[:height]
}

func hereLine(text string, selected bool, focused bool, width int, colors palette) line {
	switch {
	case !selected:
		return line{text: "  " + text}
	case !focused:
		return line{text: "> " + text}
	}
	label := truncateCells(singleLine(text), width-2)
	return line{text: colors.render(toneHere, true, "> ") + colors.render(toneStrong, true, label), raw: true}
}
