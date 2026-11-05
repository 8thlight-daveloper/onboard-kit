package tui

import "fmt"

type About struct {
	Version     string
	TemplateSet string
	Places      []AboutPlace
	Problems    []string
}

type PlaceUse int

const (
	PlaceNotUsedYet PlaceUse = iota
	PlaceRead
	PlaceWritten
	PlaceReadAndWritten
)

type AboutPlace struct {
	Name   string
	Path   string
	Chosen string
	Holds  string
	Use    PlaceUse
}

const aboutLabelWidth = 14

func (use PlaceUse) sentence() string {
	switch use {
	case PlaceRead:
		return "The tool reads it."
	case PlaceWritten:
		return "The tool writes it."
	case PlaceReadAndWritten:
		return "The tool reads and writes it."
	}
	return "Not used yet."
}

func aboutLines(about About, width int) []line {
	content := []line{strongLine("About this tool"), blankLine()}
	for _, problem := range about.Problems {
		content = append(content, lines(toneProblem, hanging("! problem: ", problem, width))...)
		content = append(content, blankLine())
	}
	content = append(content, lines(toneBody, hanging(labelled("Version"), visibleText(about.Version), width))...)
	content = append(content, lines(toneBody, hanging(labelled("Template set"), visibleText(about.TemplateSet), width))...)
	content = append(content, blankLine(), strongLine("Places this tool uses"), blankLine())
	for _, place := range about.Places {
		content = append(content, place.lines(width)...)
	}
	return content
}

func labelled(label string) string {
	return label + spaces(aboutLabelWidth-cellWidth(label))
}

func (place AboutPlace) lines(width int) []line {
	path := visibleText(place.Path)
	if place.Use == PlaceNotUsedYet && path == "" {
		path = "not used yet"
	}
	content := lines(toneStrong, hanging(labelled(place.Name), path, width))
	indent := spaces(aboutLabelWidth)
	if place.Use != PlaceNotUsedYet || place.Path != "" {
		content = append(content, lines(toneBody, hanging(indent, place.Use.sentence(), width))...)
	}
	content = append(content, lines(toneFaint, hanging(indent, place.Holds, width))...)
	if place.Chosen != "" {
		content = append(content, lines(toneFaint, hanging(indent, place.Chosen, width))...)
	}
	return append(content, blankLine())
}

func aboutScrollLimit(content []line, height int) int {
	if len(content) <= height {
		return 0
	}
	return max(len(content)-max(height-1, 1), 0)
}

func aboutWindow(content []line, offset int, height int) []line {
	if len(content) <= height {
		return content
	}
	shown := max(height-1, 1)
	offset = max(min(offset, aboutScrollLimit(content, height)), 0)
	last := min(offset+shown, len(content))
	window := append([]line{}, content[offset:last]...)
	return append(window, faintLine(fmt.Sprintf("Lines %d to %d of %d. up/down scroll.", offset+1, last, len(content))))
}

func scrolledBy(key string, offset int, height int) (int, bool) {
	switch key {
	case "up", "k":
		return offset - 1, true
	case "down", "j":
		return offset + 1, true
	case "pgup":
		return offset - max(height-1, 1), true
	case "pgdown", " ":
		return offset + max(height-1, 1), true
	case "home", "g":
		return 0, true
	case "end", "G":
		return 1 << 30, true
	}
	return offset, false
}
