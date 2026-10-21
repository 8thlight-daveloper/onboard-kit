package assembler

import (
	"strings"

	"onboard-kit/internal/model"
)

const escalationSeparator = ", then "

func escalationPart(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return "Escalation: " + strings.Join(items, escalationSeparator) + "."
}

func (a Assembler) fillEscalations(filled model.Section, original model.Section, context planContext) (model.Section, []model.Problem) {
	var problems []model.Problem
	var itemProblems []model.Problem
	filled.Escalation, itemProblems = a.fillItems(original.Escalation, original, context)
	problems = append(problems, itemProblems...)
	for index, step := range original.Steps {
		filled.Steps[index].Escalation, itemProblems = a.fillItems(step.Escalation, original, context)
		problems = append(problems, itemProblems...)
	}
	return filled, problems
}

func (a Assembler) fillItems(items []string, section model.Section, context planContext) ([]string, []model.Problem) {
	var texts []string
	var problems []model.Problem
	for _, item := range items {
		filled, itemProblems := a.resolver.Resolve(itemSection(section, item), context.answers, context.directory, context.team)
		texts = append(texts, filled.Body)
		problems = append(problems, itemProblems...)
	}
	return texts, problems
}

func itemSection(section model.Section, item string) model.Section {
	return model.Section{ID: section.ID, Variables: section.Variables, Body: item}
}
