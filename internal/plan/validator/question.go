package validator

import (
	"strings"

	"onboard-kit/internal/model"
)

const planScope = "plan"

type question struct {
	id    string
	scope string
	name  string
}

func parseQuestion(id string) question {
	scope, name, _ := strings.Cut(id, ".")
	return question{id: id, scope: scope, name: name}
}

func declaredVariable(plan model.Plan, target question) (model.Variable, bool) {
	if !strings.Contains(target.id, ".") {
		return model.Variable{}, false
	}
	variables := plan.Variables
	if target.scope != planScope {
		section, held := heldSection(plan, target.scope)
		if !held {
			return model.Variable{}, false
		}
		variables = section.Variables
	}
	for _, variable := range variables {
		if variable.Name == target.name {
			return variable, true
		}
	}
	return model.Variable{}, false
}

func heldSection(plan model.Plan, id string) (model.Section, bool) {
	for _, section := range plan.Sections {
		if section.ID == id {
			return section, true
		}
	}
	return model.Section{}, false
}

func sectionOf(plan model.Plan, target question) string {
	if target.scope == planScope {
		return ""
	}
	if _, held := heldSection(plan, target.scope); held {
		return target.scope
	}
	return ""
}
