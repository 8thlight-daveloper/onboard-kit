package writer

import (
	"strings"

	"onboard-kit/internal/model"
)

func findQuestion(plan model.Plan, question string) (string, string, model.Variable, bool) {
	scope, name, dotted := strings.Cut(question, ".")
	if !dotted || scope == "" || name == "" {
		return scope, name, model.Variable{}, false
	}
	variables := plan.Variables
	if scope != "plan" {
		var found bool
		variables, found = sectionVariables(plan.Sections, scope)
		if !found {
			return scope, name, model.Variable{}, false
		}
	}
	variable, found := exactVariable(variables, name)
	return scope, name, variable, found
}

func sectionVariables(sections []model.Section, scope string) ([]model.Variable, bool) {
	var variables []model.Variable
	count := 0
	for _, section := range sections {
		if section.ID == scope {
			variables = section.Variables
			count++
		}
	}
	return variables, count == 1
}

func exactVariable(variables []model.Variable, name string) (model.Variable, bool) {
	var selected model.Variable
	count := 0
	for _, variable := range variables {
		if variable.Name == name {
			selected = variable
			count++
		}
	}
	return selected, count == 1
}
