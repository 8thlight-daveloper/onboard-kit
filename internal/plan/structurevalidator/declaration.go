package structurevalidator

import (
	"sort"

	"onboard-kit/internal/model"
)

type declaration struct {
	parent    string
	requires  []string
	after     []string
	bootstrap bool
	timing    model.Timing
}

func declarationsOf(sections []model.Section) (map[string]declaration, []issue) {
	declarations := map[string]declaration{}
	duplicates := map[string]bool{}
	for _, section := range sections {
		if _, present := declarations[section.ID]; present {
			duplicates[section.ID] = true
			continue
		}
		declarations[section.ID] = declaration{
			parent: section.Parent, requires: section.Requires, after: section.After,
			bootstrap: section.Bootstrap, timing: section.Timing,
		}
	}
	var issues []issue
	for identity := range duplicates {
		issues = append(issues, issue{
			code: "plan.duplicate-section-id", section: identity, item: -1,
		})
	}
	return declarations, issues
}

func sortedIdentities(declarations map[string]declaration) []string {
	identities := make([]string, 0, len(declarations))
	for identity := range declarations {
		identities = append(identities, identity)
	}
	sort.Strings(identities)
	return identities
}
