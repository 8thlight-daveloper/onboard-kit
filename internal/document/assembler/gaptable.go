package assembler

import (
	"strings"

	"onboard-kit/internal/model"
)

const gapTableHeader = "## Gaps to fill\n\n| Gap | Who fills it | By when |\n| --- | --- | --- |"

func gapTable(sections []filledSection) string {
	lines := []string{gapTableHeader}
	for _, section := range sections {
		for _, problem := range section.problems {
			lines = append(lines, gapRow(section, problem))
		}
	}
	return strings.Join(lines, lineBreak)
}

func gapRow(section filledSection, problem model.Problem) string {
	owner := prose{blanks: onboardingBlanks(section.declared)}.of(section.filled.Owner)
	cells := []string{
		gapDescription(section.declared, problem),
		owner,
		capitalized(timingInWords(section.filled.Timing)),
	}
	return "| " + strings.Join(escapedCells(cells), " | ") + " |"
}

func gapDescription(declared []model.Variable, problem model.Problem) string {
	name := strings.TrimPrefix(problem.Question, problem.Section+".")
	for _, variable := range declared {
		if variable.Name == name && !isBlank(variable.Description) {
			return variable.Description
		}
	}
	return name
}
