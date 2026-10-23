package structurevalidator

import "onboard-kit/internal/model"

type Validator struct{}

func New() Validator {
	return Validator{}
}

func (Validator) Validate(plan model.Plan) []model.Problem {
	declarations, duplicates := declarationsOf(plan.Sections)
	if len(duplicates) != 0 {
		return problemsOf(duplicates)
	}
	identities := sortedIdentities(declarations)
	parents, parentIssues := parentRelations(declarations, identities)
	requires, requiresIssues := prerequisiteRelations(declarations, identities)
	groups := timingGroups(declarations, identities)
	after, afterIssues := legacyRelations(declarations, identities, groups)

	issues := append(parentIssues, requiresIssues...)
	issues = append(issues, afterIssues...)
	for _, members := range parents.cyclicComponents() {
		issues = append(issues, cycleIssue("parent", members))
	}
	pureCycleMembers := map[string]bool{}
	for _, pair := range []struct {
		field string
		graph relation
	}{{"requires", requires}, {"after", after}} {
		for _, members := range pair.graph.cyclicComponents() {
			issues = append(issues, cycleIssue(pair.field, members))
			for _, member := range members {
				pureCycleMembers[member] = true
			}
		}
	}
	for _, members := range unionOf(requires, after).cyclicComponents() {
		containsPureCycle := false
		for _, member := range members {
			containsPureCycle = containsPureCycle || pureCycleMembers[member]
		}
		if !containsPureCycle {
			issues = append(issues, issue{
				code: "plan.ordering-conflict", section: members[0],
				field: "ordering", item: -1, members: members,
			})
		}
	}
	return problemsOf(issues)
}
