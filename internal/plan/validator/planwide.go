package validator

import (
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

var planWideSortKeys = map[string][]string{
	"directory.contact-unreachable":  {"contact"},
	"directory.email-malformed":      {"contact"},
	"directory.duplicate-id":         {"kind", "id"},
	"directory.holder-not-a-contact": {"role", "team", "holder", "team-contact"},
	"directory.role-vacant":          {"role"},
	"directory.team-role-unknown":    {"team", "role"},
	"plan.duplicate-section-id":      {"section"},
}

func planWideProblems(plan model.Plan) []model.Problem {
	problems := duplicateIDProblems(plan.Directory)
	problems = append(problems, holderProblems(plan.Directory)...)
	problems = append(problems, contactProblems(plan.Directory)...)
	problems = append(problems, duplicateSectionProblems(plan)...)
	problems = append(problems, duplicateTeamVariableProblems(plan)...)
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
