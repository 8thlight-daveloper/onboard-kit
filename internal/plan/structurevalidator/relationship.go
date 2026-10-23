package structurevalidator

func parentRelations(declarations map[string]declaration, identities []string) (relation, []issue) {
	parents := newRelation(identities)
	var issues []issue
	for _, identity := range identities {
		parent := declarations[identity].parent
		if parent == "" {
			continue
		}
		code := ""
		if parent == identity {
			code = "plan.parent-self"
		} else if _, present := declarations[parent]; !present {
			code = "plan.parent-unknown"
		}
		if code != "" {
			issues = append(issues, issue{code: code, section: identity, field: "parent", item: -1})
		} else {
			parents.add(identity, parent)
		}
	}
	return parents, issues
}

func prerequisiteRelations(declarations map[string]declaration, identities []string) (relation, []issue) {
	requires := newRelation(identities)
	var issues []issue
	for _, identity := range identities {
		holder := declarations[identity]
		seen := map[string]bool{}
		for index, target := range holder.requires {
			if seen[target] {
				continue
			}
			seen[target] = true
			prerequisite, present := declarations[target]
			switch {
			case target == identity:
				issues = append(issues, referenceIssue("plan.requires-self", identity, "requires", index))
			case !present:
				issues = append(issues, referenceIssue("plan.requires-unknown", identity, "requires", index))
			default:
				requires.add(target, identity)
				if holder.bootstrap && !prerequisite.bootstrap {
					issues = append(issues, referenceIssue("plan.bootstrap-requires-nonbootstrap", identity, "requires", index))
				}
			}
		}
	}
	return requires, issues
}

func legacyRelations(declarations map[string]declaration, identities []string, groups map[string]timingGroup) (relation, []issue) {
	after := newRelation(identities)
	var issues []issue
	for _, identity := range identities {
		holder := declarations[identity]
		seen := map[string]bool{}
		for index, target := range holder.after {
			if seen[target] {
				continue
			}
			seen[target] = true
			predecessor, present := declarations[target]
			switch {
			case target == identity:
				issues = append(issues, referenceIssue("plan.after-self", identity, "after", index))
			case !present:
				issues = append(issues, referenceIssue("plan.after-unknown", identity, "after", index))
			case groups[identity].valid && groups[target].valid && holder.parent == predecessor.parent && groups[identity].timing == groups[target].timing:
				after.add(target, identity)
				if holder.bootstrap && !predecessor.bootstrap {
					issues = append(issues, referenceIssue("plan.bootstrap-after-nonbootstrap", identity, "after", index))
				}
			}
		}
	}
	return after, issues
}
