package structurevalidator

import (
	"cmp"
	"encoding/json"
	"slices"
	"sort"
	"strconv"

	"onboard-kit/internal/model"
)

type issue struct {
	code    string
	section string
	field   string
	item    int
	members []string
}

func referenceIssue(code, section, field string, item int) issue {
	return issue{code: code, section: section, field: field, item: item}
}

func cycleIssue(field string, members []string) issue {
	return issue{
		code: "plan." + field + "-cycle", section: members[0], field: field, item: -1, members: members,
	}
}

func problemsOf(issues []issue) []model.Problem {
	sort.Slice(issues, func(left, right int) bool {
		return compareIssues(issues[left], issues[right]) < 0
	})
	var problems []model.Problem
	for index, found := range issues {
		if index > 0 && compareIssues(issues[index-1], found) == 0 {
			continue
		}
		detail := map[string]string{}
		if found.field != "" {
			detail["field"] = found.field
		}
		if found.item >= 0 {
			detail["item"] = strconv.Itoa(found.item)
		}
		if len(found.members) != 0 {
			members, _ := json.Marshal(found.members)
			detail["members"] = string(members)
		}
		problems = append(problems, model.Problem{
			Code: found.code, Severity: model.SeverityError, Message: repairMessage(found.code),
			Section: found.section, Detail: detail,
		})
	}
	return problems
}

func compareIssues(left, right issue) int {
	for _, pair := range [][2]string{{left.code, right.code}, {left.section, right.section}, {left.field, right.field}} {
		if order := cmp.Compare(pair[0], pair[1]); order != 0 {
			return order
		}
	}
	if order := cmp.Compare(left.item, right.item); order != 0 {
		return order
	}
	return slices.Compare(left.members, right.members)
}

func repairMessage(code string) string {
	switch code {
	case "plan.duplicate-section-id":
		return "Give each section a unique ID before checking relationships."
	case "plan.parent-unknown":
		return "Choose an existing section as parent, or remove parent."
	case "plan.parent-self":
		return "Remove the section's reference to itself as parent."
	case "plan.parent-cycle":
		return "Remove a parent link within the listed cycle members."
	case "plan.requires-unknown":
		return "Choose an existing prerequisite, or remove this requires item."
	case "plan.requires-self":
		return "Remove the section's reference to itself in requires."
	case "plan.requires-cycle":
		return "Remove a requires edge within the listed cycle members."
	case "plan.bootstrap-requires-nonbootstrap":
		return "Make this prerequisite bootstrap, or correct the bootstrap claim."
	case "plan.after-unknown":
		return "Choose an existing section, or remove this after item."
	case "plan.after-self":
		return "Remove the section's reference to itself in after."
	case "plan.after-cycle":
		return "Remove an after edge within the listed cycle members."
	case "plan.bootstrap-after-nonbootstrap":
		return "Remove this ordering hint, or correct the bootstrap claim."
	case "plan.ordering-conflict":
		return "Change a requires or effective after edge within these members so both orders can coexist."
	}
	return ""
}
