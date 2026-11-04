package tui

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

func cellWidth(text string) int { return lipgloss.Width(text) }

func spaces(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(" ", count)
}

func singleLine(text string) string {
	return strings.NewReplacer("\t", " ", "\r", "", "\n", " ").Replace(text)
}

func splitCells(text string, width int) (string, string) {
	used := 0
	for index, character := range text {
		size := cellWidth(string(character))
		if used+size > width && index > 0 {
			return text[:index], text[index:]
		}
		used += size
	}
	return text, ""
}

func truncateCells(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if cellWidth(text) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}
	head, _ := splitCells(text, width-1)
	return head + "…"
}

func wrapText(text string, width int) []string {
	width = max(width, 1)
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		lines = append(lines, wrapParagraph(paragraph, width)...)
	}
	return lines
}

func wrapParagraph(paragraph string, width int) []string {
	words := strings.Fields(paragraph)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	current := ""
	for _, word := range words {
		for cellWidth(word) > width {
			head, rest := splitCells(word, width)
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, head)
			word = rest
		}
		switch {
		case current == "":
			current = word
		case cellWidth(current)+1+cellWidth(word) <= width:
			current += " " + word
		default:
			lines = append(lines, current)
			current = word
		}
	}
	return append(lines, current)
}

func wrapIndented(text string, width int) []string {
	trimmed := strings.TrimLeft(text, " ")
	indent := len(text) - len(trimmed)
	if indent > width/2 {
		indent = 0
	}
	var lines []string
	for _, piece := range wrapText(trimmed, width-indent) {
		lines = append(lines, spaces(indent)+piece)
	}
	return lines
}

func hanging(prefix string, text string, width int) []string {
	indent := cellWidth(prefix)
	var lines []string
	for index, piece := range wrapText(text, width-indent) {
		lead := spaces(indent)
		if index == 0 {
			lead = prefix
		}
		lines = append(lines, lead+piece)
	}
	return lines
}

func visibleText(text string) string {
	var shown strings.Builder
	for _, character := range text {
		if unicode.IsControl(character) && character != '\t' && character != '\n' {
			quoted := strconv.QuoteRuneToASCII(character)
			shown.WriteString(quoted[1 : len(quoted)-1])
			continue
		}
		shown.WriteRune(character)
	}
	return shown.String()
}
