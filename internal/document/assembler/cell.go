package assembler

import "strings"

var cellEscaper = strings.NewReplacer("|", `\|`, "\n", wordSeparator)

func escapedCells(cells []string) []string {
	escaped := make([]string, len(cells))
	for index, cell := range cells {
		escaped[index] = cellEscaper.Replace(trimmed(cell))
	}
	return escaped
}
