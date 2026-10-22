package validator

import "onboard-kit/internal/model"

func duplicateVariableProblems(scope, section string, variables []model.Variable) []model.Problem {
	counts := map[string]int{}
	var problems []model.Problem
	for _, variable := range variables {
		counts[variable.Name]++
		if counts[variable.Name] == 2 {
			problems = append(problems, model.Problem{
				Code:     "plan.duplicate-variable-name",
				Severity: model.SeverityError,
				Message:  "Variable " + variable.Name + " is declared more than once in scope " + scope + ". Give each variable a unique name.",
				Section:  section,
				Detail:   map[string]string{"scope": scope, "name": variable.Name},
			})
		}
	}
	return problems
}
