package authoringvalidator

import (
	"fmt"
	"reflect"
	"strings"

	"onboard-kit/internal/model"
)

type resolvedClause struct {
	target  int
	test    model.AnswerTest
	section string
	enabled bool
}

func (a *analysis) ancestors(section int) ([]int, bool) {
	if a.duplicateSections {
		return nil, false
	}
	seen := map[int]bool{section: true}
	var result []int
	for parent := a.plan.Sections[section].Parent; parent != ""; {
		indices := a.sections[parent]
		if len(indices) != 1 {
			return nil, false
		}
		index := indices[0]
		if seen[index] {
			return nil, false
		}
		seen[index] = true
		result = append(result, index)
		parent = a.plan.Sections[index].Parent
	}
	return result, true
}

func (a *analysis) resolve(section int, reference string) (int, bool) {
	scope := section
	name := reference
	if before, after, qualified := strings.Cut(reference, "."); qualified {
		name = after
		if before == "plan" {
			scope = -1
		} else {
			if a.duplicateSections {
				return -1, true
			}
			indices := a.sections[before]
			if len(indices) == 0 {
				return -1, false
			}
			if len(indices) > 1 {
				return -1, true
			}
			scope = indices[0]
		}
	}
	if scope >= 0 && len(a.sections[a.plan.Sections[scope].ID]) > 1 {
		return -1, true
	}
	indices := a.scopes[scope+1][name]
	if len(indices) == 0 {
		return -1, false
	}
	if len(indices) > 1 {
		return -1, true
	}
	return indices[0], false
}

func (a *analysis) analyzeConditions() {
	for index, s := range a.plan.Sections {
		path := fmt.Sprintf("sections[%d].includeIf", index)
		targets := a.checkCondition(s.IncludeIf, index, path, a.sectionLocation(index), true)
		for target := range targets {
			addEdge(a.includeEdges, index, target)
			a.used[target] = true
		}
	}
	for index, d := range a.declarations {
		if d.section < 0 {
			continue
		}
		targets := a.checkCondition(d.variable.AskIf, d.section, d.path+".askIf", d, false)
		if d.unique {
			for target := range targets {
				addEdge(a.askEdges, index, target)
			}
		}
	}
}

func addEdge(graph map[int]map[int]bool, source, target int) {
	if graph[source] == nil {
		graph[source] = map[int]bool{}
	}
	graph[source][target] = true
}

func (a *analysis) checkCondition(condition *model.Condition, section int, path string, d declaration, include bool) map[int]bool {
	targets := map[int]bool{}
	if condition == nil {
		return targets
	}
	if len(condition.All) == 0 {
		a.emit("condition-empty", path, d)
		return targets
	}
	var clauses []resolvedClause
	for i, clause := range condition.All {
		clausePath := fmt.Sprintf("%s.all[%d]", path, i)
		question := clause.Question != ""
		switchClause := clause.Section != ""
		if question == switchClause || (switchClause && !reflect.DeepEqual(clause.Answer, model.AnswerTest{})) {
			a.emit("clause", clausePath, d)
			continue
		}
		if switchClause {
			indices := a.sections[clause.Section]
			if len(indices) == 0 {
				a.emit("section-reference", clausePath+".section", d)
			} else if len(indices) == 1 {
				clauses = append(clauses, resolvedClause{target: -1, section: clause.Section, enabled: clause.Enabled})
			}
			continue
		}
		target, ambiguous := a.resolve(section, clause.Question)
		if ambiguous {
			continue
		}
		if target < 0 {
			a.emit("question-reference", clausePath+".question", d)
			continue
		}
		if include && a.declarations[target].section >= 0 {
			if a.declarations[target].section == section {
				a.emit("include-scope", clausePath+".question", d)
				continue
			}
			ancestors, known := a.ancestors(section)
			if !known {
				continue
			}
			allowed := false
			for _, ancestor := range ancestors {
				if a.declarations[target].section == ancestor {
					allowed = true
				}
			}
			if !allowed {
				a.emit("include-scope", clausePath+".question", d)
				continue
			}
		}
		targets[target] = true
		if a.checkPredicate(clause.Answer, a.declarations[target].variable, clausePath+".answer", d) {
			clauses = append(clauses, resolvedClause{target: target, test: clause.Answer})
		}
	}
	if a.contradiction(clauses) {
		a.emit("contradiction", path, d)
	}
	return targets
}

func testCount(test model.AnswerTest) int {
	count := 0
	if test.Is.Kind != model.ValueKindNone {
		count++
	}
	if test.IsNot.Kind != model.ValueKindNone {
		count++
	}
	if test.IsOneOf != nil {
		count++
	}
	if test.IsNotOneOf != nil {
		count++
	}
	if test.IsSettled != nil {
		count++
	}
	return count
}

func (a *analysis) checkPredicate(test model.AnswerTest, v model.Variable, path string, d declaration) bool {
	count := testCount(test)
	if count == 0 {
		a.emit("predicate", path, d)
		return false
	}
	if !validType(v) {
		return false
	}
	comparisons := test.Is.Kind != model.ValueKindNone || test.IsNot.Kind != model.ValueKindNone || test.IsOneOf != nil || test.IsNotOneOf != nil
	if listType(v) && comparisons {
		a.emit("predicate", path, d)
		return false
	}
	valid := true
	for _, single := range []struct {
		name  string
		value model.Value
	}{{"is", test.Is}, {"isNot", test.IsNot}} {
		if single.value.Kind != model.ValueKindNone && !a.validValue(v, single.value, true) {
			a.emit("predicate-value", path+"."+single.name, d)
			valid = false
		}
	}
	for _, list := range []struct {
		name   string
		values []model.Value
	}{{"isOneOf", test.IsOneOf}, {"isNotOneOf", test.IsNotOneOf}} {
		if list.values == nil {
			continue
		}
		badList := len(list.values) < 2 || variableType(v) == model.VariableTypeBoolean
		if variableType(v) == model.VariableTypeChoice && validOptions(v) {
			all := true
			for _, option := range v.Options {
				found := false
				for _, value := range list.values {
					if value.Kind == model.ValueKindText && value.Text == option {
						found = true
					}
				}
				if !found {
					all = false
				}
			}
			if all {
				badList = true
			}
		}
		if badList {
			a.emit("predicate-list", path+"."+list.name, d)
			valid = false
			continue
		}
		for i, value := range list.values {
			if !a.validValue(v, value, true) {
				a.emit("predicate-value", fmt.Sprintf("%s.%s[%d]", path, list.name, i), d)
				valid = false
			}
		}
	}
	return valid
}

type scalar struct {
	kind model.ValueKind
	text string
	flag bool
}

func scalarValue(value model.Value) scalar {
	result := scalar{kind: value.Kind}
	if value.Kind == model.ValueKindText {
		result.text = value.Text
	}
	if value.Kind == model.ValueKindFlag {
		result.flag = value.Flag
	}
	return result
}

func valueSet(values []model.Value) map[scalar]bool {
	result := map[scalar]bool{}
	for _, value := range values {
		result[scalarValue(value)] = true
	}
	return result
}

func (a *analysis) contradiction(clauses []resolvedClause) bool {
	groups := map[int][]model.AnswerTest{}
	switches := map[string]bool{}
	for _, clause := range clauses {
		if clause.section != "" {
			if previous, ok := switches[clause.section]; ok && previous != clause.enabled {
				return true
			}
			switches[clause.section] = clause.enabled
		} else {
			groups[clause.target] = append(groups[clause.target], clause.test)
		}
	}
	for target, tests := range groups {
		if contradictoryTests(a.declarations[target].variable, tests) {
			return true
		}
	}
	return false
}

func contradictoryTests(variable model.Variable, tests []model.AnswerTest) bool {
	var positive map[scalar]bool
	excluded := map[scalar]bool{}
	settledTrue, settledFalse, comparison := false, false, false
	intersect := func(next map[scalar]bool) {
		if positive == nil {
			positive = next
			return
		}
		for value := range positive {
			if !next[value] {
				delete(positive, value)
			}
		}
	}
	for _, test := range tests {
		if test.IsSettled != nil {
			if *test.IsSettled {
				settledTrue = true
			} else {
				settledFalse = true
			}
		}
		if test.Is.Kind != model.ValueKindNone {
			comparison = true
			intersect(map[scalar]bool{scalarValue(test.Is): true})
		}
		if test.IsNot.Kind != model.ValueKindNone {
			comparison = true
			excluded[scalarValue(test.IsNot)] = true
		}
		if test.IsOneOf != nil {
			comparison = true
			intersect(valueSet(test.IsOneOf))
		}
		if test.IsNotOneOf != nil {
			comparison = true
			for value := range valueSet(test.IsNotOneOf) {
				excluded[value] = true
			}
		}
	}
	if settledFalse && (settledTrue || comparison) {
		return true
	}
	if positive != nil {
		for value := range excluded {
			delete(positive, value)
		}
		if len(positive) == 0 {
			return true
		}
	}
	if variableType(variable) == model.VariableTypeBoolean {
		return excluded[scalar{kind: model.ValueKindFlag, flag: true}] && excluded[scalar{kind: model.ValueKindFlag, flag: false}]
	}
	if variableType(variable) == model.VariableTypeChoice && validOptions(variable) {
		for _, option := range variable.Options {
			if !excluded[scalar{kind: model.ValueKindText, text: option}] {
				return false
			}
		}
		return true
	}
	return false
}
