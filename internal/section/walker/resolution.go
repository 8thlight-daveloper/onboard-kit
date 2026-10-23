package walker

import (
	"sort"

	"onboard-kit/internal/model"
	"onboard-kit/internal/section/resolver"
)

func teamIdentifier(plan model.Plan) string {
	for _, variable := range plan.Variables {
		if variable.Type != model.VariableTypeTeam {
			continue
		}
		answer := plan.Answers.Plan[variable.Name]
		if answer.Kind == model.ValueKindText && answer.Text != "" {
			for _, team := range plan.Directory.Teams {
				if team.ID == answer.Text {
					return team.ID
				}
			}
		}
		return ""
	}
	return ""
}

func resolveEscalation(section model.Section, plan model.Plan, team string) ([]string, []model.Problem) {
	var rendered []string
	var problems []model.Problem
	for _, item := range section.Escalation {
		carrier := model.Section{ID: section.ID, Variables: section.Variables, Body: item}
		resolved, gaps := resolver.New().Resolve(carrier, plan.Answers, plan.Directory, team)
		rendered = append(rendered, resolved.Body)
		problems = append(problems, gaps...)
	}
	return rendered, problems
}

func sectionProblems(section model.Section, problems []model.Problem) []model.Problem {
	type identity struct {
		question string
		code     string
	}
	seen := make(map[identity]bool, len(problems))
	var unique []model.Problem
	for _, problem := range problems {
		key := identity{question: problem.Question, code: problem.Code}
		if !seen[key] {
			unique = append(unique, problem)
			seen[key] = true
		}
	}
	ranks := make(map[string]int, len(section.Variables))
	for rank, variable := range section.Variables {
		ranks[section.ID+"."+variable.Name] = rank
	}
	sort.Slice(unique, func(i, j int) bool {
		if ranks[unique[i].Question] != ranks[unique[j].Question] {
			return ranks[unique[i].Question] < ranks[unique[j].Question]
		}
		return unique[i].Code < unique[j].Code
	})
	return unique
}
