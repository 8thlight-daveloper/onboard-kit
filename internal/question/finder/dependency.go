package finder

import "strings"

func namedIDs(asked candidate) []string {
	if asked.variable.AskIf == nil {
		return nil
	}
	ids := []string{}
	for _, clause := range asked.variable.AskIf.All {
		if clause.Question == "" {
			continue
		}
		if id := qualify(clause.Question, asked.scope); id != asked.id() {
			ids = append(ids, id)
		}
	}
	return ids
}

func qualify(question string, scope string) string {
	if strings.Contains(question, ".") {
		return question
	}
	return scope + "." + question
}

func positionsNamed(candidates []candidate) [][]int {
	positions := map[string]int{}
	for position, asked := range candidates {
		if _, seen := positions[asked.id()]; !seen {
			positions[asked.id()] = position
		}
	}
	named := make([][]int, len(candidates))
	for position, asked := range candidates {
		for _, id := range namedIDs(asked) {
			if target, found := positions[id]; found && target != position {
				named[position] = append(named[position], target)
			}
		}
	}
	return named
}
