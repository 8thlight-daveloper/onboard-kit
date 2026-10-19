package instantiator

import (
	"sort"
	"strconv"

	"onboard-kit/internal/model"
)

func duplicateIDProblems(templates []model.Section) []model.Problem {
	counts := map[string]int{}
	for _, template := range templates {
		counts[template.ID]++
	}
	var repeated []string
	for id, count := range counts {
		if count > 1 {
			repeated = append(repeated, id)
		}
	}
	sort.Strings(repeated)
	problems := make([]model.Problem, len(repeated))
	for index, id := range repeated {
		problems[index] = model.Problem{
			Code:     "template.duplicate-id",
			Severity: model.SeverityError,
			Message:  "more than one template has the id " + strconv.Quote(id),
			Section:  id,
			Detail:   map[string]string{"count": strconv.Itoa(counts[id])},
		}
	}
	return problems
}

func namelessVariableProblems(variables []model.Variable) []model.Problem {
	var problems []model.Problem
	for index, variable := range variables {
		if variable.Name != "" {
			continue
		}
		problems = append(problems, model.Problem{
			Code:     "variable.no-name",
			Severity: model.SeverityError,
			Message:  "an axis variable has no name",
			Detail:   map[string]string{"index": strconv.Itoa(index)},
		})
	}
	return problems
}
