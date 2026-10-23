package validator

import (
	"slices"
	"strings"

	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/structurevalidator"
)

var planWideSortKeys = map[string][]string{
	"directory.contact-unreachable":  {"contact"},
	"directory.email-malformed":      {"contact"},
	"directory.duplicate-id":         {"kind", "id"},
	"directory.holder-not-a-contact": {"role", "team", "holder", "team-contact"},
	"directory.role-vacant":          {"role"},
	"directory.team-role-unknown":    {"team", "role"},
	"plan.duplicate-section-id":      {"section"},
	"plan.duplicate-variable-name":   {"scope", "name"},
	"plan.duplicate-step-id":         {"section", "step"},
}

func planWideProblems(plan model.Plan) []model.Problem {
	problems := duplicateIDProblems(plan.Directory)
	problems = append(problems, holderProblems(plan.Directory)...)
	problems = append(problems, contactProblems(plan.Directory)...)
	problems = append(problems, structurevalidator.New().Validate(plan)...)
	problems = append(problems, duplicateTeamVariableProblems(plan)...)
	problems = append(problems, duplicateVariableProblems(planScope, "", plan.Variables)...)
	for _, section := range plan.Sections {
		problems = append(problems, duplicateVariableProblems(section.ID, section.ID, section.Variables)...)
		problems = append(problems, duplicateStepProblems(section)...)
	}
	slices.SortStableFunc(problems, comparePlanWide)
	return problems
}

func comparePlanWide(left, right model.Problem) int {
	if byCode := strings.Compare(left.Code, right.Code); byCode != 0 {
		return byCode
	}
	for _, key := range planWideSortKeys[left.Code] {
		if byKey := strings.Compare(sortValue(left, key), sortValue(right, key)); byKey != 0 {
			return byKey
		}
	}
	return 0
}

func sortValue(problem model.Problem, key string) string {
	if key == "section" {
		return problem.Section
	}
	return problem.Detail[key]
}
