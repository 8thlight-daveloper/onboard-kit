package structurevalidator

import "onboard-kit/internal/model"

type timingGroup struct {
	timing model.Timing
	valid  bool
}

type timingSearch struct {
	declarations map[string]declaration
	groups       map[string]timingGroup
	visiting     map[string]bool
}

func timingGroups(declarations map[string]declaration, identities []string) map[string]timingGroup {
	search := timingSearch{
		declarations: declarations,
		groups:       map[string]timingGroup{},
		visiting:     map[string]bool{},
	}
	for _, identity := range identities {
		search.groupOf(identity)
	}
	return search.groups
}

func (search timingSearch) groupOf(identity string) timingGroup {
	if group, present := search.groups[identity]; present {
		return group
	}
	if search.visiting[identity] {
		return timingGroup{}
	}
	search.visiting[identity] = true
	held := search.declarations[identity]
	group := timingGroup{timing: held.timing, valid: true}
	if held.parent == "" {
		if group.timing == "" {
			group.timing = model.DefaultTiming
		}
	} else if _, present := search.declarations[held.parent]; !present {
		group.valid = false
	} else {
		parent := search.groupOf(held.parent)
		group.valid = parent.valid
		if group.timing == "" {
			group.timing = parent.timing
		}
	}
	search.visiting[identity] = false
	search.groups[identity] = group
	return group
}
