package tui

import "strings"

type line struct {
	text string
	tone tone
	raw  bool
}

func plainLine(text string) line   { return line{text: text} }
func mutedLine(text string) line   { return line{text: text, tone: toneMuted} }
func headingLine(text string) line { return line{text: text, tone: toneHeading} }

func lines(tone tone, texts []string) []line {
	result := make([]line, 0, len(texts))
	for _, text := range texts {
		result = append(result, line{text: text, tone: tone})
	}
	return result
}

func renderCell(entry line, width int, p palette) string {
	if entry.raw {
		return entry.text + spaces(width-cellWidth(entry.text))
	}
	text := truncateCells(singleLine(entry.text), width)
	return p.style(entry.tone).Render(text) + spaces(width-cellWidth(text))
}

func renderRow(entry line, width int, p palette) string {
	return strings.TrimRight(renderCell(entry, width, p), " ")
}

func sideBySide(left []line, right []line, l layout, p palette) []string {
	rows := make([]string, 0, l.bodyHeight)
	for index := 0; index < l.bodyHeight; index++ {
		var leftLine, rightLine line
		if index < len(left) {
			leftLine = left[index]
		}
		if index < len(right) {
			rightLine = right[index]
		}
		rows = append(rows, strings.TrimRight(renderCell(leftLine, l.listWidth, p)+" │ "+renderCell(rightLine, l.mainWidth, p), " "))
	}
	return rows
}

func singlePane(content []line, l layout, p palette) []string {
	rows := make([]string, 0, l.bodyHeight)
	for index := 0; index < l.bodyHeight; index++ {
		if index < len(content) {
			rows = append(rows, renderRow(content[index], l.width, p))
			continue
		}
		rows = append(rows, "")
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
