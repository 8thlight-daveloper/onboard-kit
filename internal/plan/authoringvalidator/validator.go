package authoringvalidator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	contract "onboard-kit/contracts/plan/authoringvalidator"
	"onboard-kit/internal/model"
	"onboard-kit/internal/secrets"
)

type Validator struct {
	checker secrets.Checker
}

var _ contract.Validator = Validator{}

func New() Validator { return Validator{checker: secrets.New()} }

type declaration struct {
	variable  model.Variable
	section   int
	path      string
	question  string
	sectionID string
	unique    bool
}

type analysis struct {
	checker           secrets.Checker
	plan              model.Plan
	problems          []model.Problem
	sections          map[string][]int
	scopes            []map[string][]int
	declarations      []declaration
	duplicateSections bool
	used              map[int]bool
	edges             map[int]map[int]bool
	askEdges          map[int]map[int]bool
	includeEdges      map[int]map[int]bool
}

func repairMessage(code string) string {
	return map[string]string{
		"section-id":         "Use a nonreserved lowercase section slug.",
		"step-id":            "Use a lowercase step slug or omit the identifier.",
		"required-prose":     "Supply nonblank required section prose.",
		"variable-name":      "Use a nonreserved variable name with at least three characters.",
		"variable-type":      "Choose a supported variable type.",
		"variable-audience":  "Choose onboarder or onboardee audience.",
		"multiple":           "Allow multiple values only for contact or role variables.",
		"choice-options":     "Supply at least one nonempty choice option.",
		"pattern":            "Use a valid RE2 string pattern.",
		"default":            "Supply a default that fits its declaration.",
		"timing":             "Choose a supported timing goal.",
		"directory-id":       "Use a lowercase directory identifier.",
		"directory-name":     "Supply a nonblank directory display name.",
		"dollar":             "Escape the dollar sign or use a declared local variable.",
		"owner-type":         "Reference a contact or role variable in owner text.",
		"escalation":         "Use one local contact or role variable reference.",
		"environment-name":   "Use an uppercase environment variable name.",
		"check":              "Supply exactly one meaningful check operation.",
		"plan-ask-if":        "Remove askIf from the plan variable.",
		"condition-empty":    "Remove the empty condition or supply its clauses.",
		"clause":             "Supply exactly one question or section-switch clause.",
		"question-reference": "Reference an existing unambiguous declared question.",
		"section-reference":  "Reference an existing declared section switch.",
		"include-scope":      "Gate a section only on plan or ancestor questions.",
		"predicate":          "Supply a supported meaningful answer predicate.",
		"predicate-value":    "Use a settled scalar value accepted by the referenced question.",
		"predicate-list":     "Use at least two values without a boolean or all-options list test.",
		"contradiction":      "Remove mutually incompatible condition tests.",
		"condition-cycle":    "Remove the question eligibility dependency cycle.",
		"unused-onboardee":   "Use this onboardee question in section content or eligibility.",
	}[code]
}

func slugValid(value string) bool {
	if value == "" || !lowerDigit(value[0]) {
		return false
	}
	for i := 1; i < len(value); i++ {
		if !lowerDigit(value[i]) && value[i] != '-' {
			return false
		}
	}
	return true
}

func lowerDigit(value byte) bool { return value >= 'a' && value <= 'z' || value >= '0' && value <= '9' }

func nameValid(value string) bool {
	if len(value) < 3 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for i := 1; i < len(value); i++ {
		if !nameCharacter(value[i]) {
			return false
		}
	}
	return true
}

func environmentValid(value string) bool {
	if value == "" || value[0] < 'A' || value[0] > 'Z' {
		return false
	}
	for i := 1; i < len(value); i++ {
		if !(value[i] >= 'A' && value[i] <= 'Z' || value[i] >= '0' && value[i] <= '9' || value[i] == '_') {
			return false
		}
	}
	return true
}

func legalSection(id string) bool {
	if !slugValid(id) {
		return false
	}
	switch id {
	case "plan", "yes", "no", "y", "n", "true", "false", "on", "off", "null":
		return false
	}
	return true
}

func legalName(name string) bool {
	if !nameValid(name) {
		return false
	}
	switch name {
	case "enabled", "yes", "true", "false", "off", "null":
		return false
	}
	return true
}

func variableType(v model.Variable) model.VariableType {
	if v.Type == "" {
		return model.VariableTypeString
	}
	return v.Type
}

func validType(v model.Variable) bool {
	switch variableType(v) {
	case model.VariableTypeString, model.VariableTypeBoolean, model.VariableTypeChoice, model.VariableTypeCommand, model.VariableTypeContact, model.VariableTypeRole, model.VariableTypeTeam:
		return true
	}
	return false
}

func listType(v model.Variable) bool {
	return variableType(v) == model.VariableTypeCommand || ((variableType(v) == model.VariableTypeContact || variableType(v) == model.VariableTypeRole) && v.Multiple)
}

func (validator Validator) Validate(plan model.Plan) []model.Problem {
	a := analysis{checker: validator.checker, plan: plan, sections: map[string][]int{}, scopes: make([]map[string][]int, len(plan.Sections)+1), used: map[int]bool{}, edges: map[int]map[int]bool{}, askEdges: map[int]map[int]bool{}, includeEdges: map[int]map[int]bool{}}
	for i, s := range plan.Sections {
		a.sections[s.ID] = append(a.sections[s.ID], i)
	}
	for _, indices := range a.sections {
		if len(indices) > 1 {
			a.duplicateSections = true
		}
	}
	a.addDeclarations(-1, plan.Variables, "variables")
	for i, s := range plan.Sections {
		a.addDeclarations(i, s.Variables, fmt.Sprintf("sections[%d].variables", i))
	}
	for i := range a.declarations {
		d := &a.declarations[i]
		d.unique = len(a.scopes[d.section+1][d.variable.Name]) == 1 && (d.section < 0 || len(a.sections[plan.Sections[d.section].ID]) == 1)
		a.checkVariable(i)
	}
	a.checkDirectory()
	for i := range plan.Sections {
		a.checkSection(i)
	}
	a.analyzeConditions()
	a.finishGraph()
	sort.SliceStable(a.problems, func(i, j int) bool {
		left, right := a.problems[i], a.problems[j]
		for _, pair := range [][2]string{{left.Detail["path"], right.Detail["path"]}, {left.Code, right.Code}, {left.Question, right.Question}, {left.Section, right.Section}, {string(left.Severity), string(right.Severity)}} {
			if pair[0] != pair[1] {
				return pair[0] < pair[1]
			}
		}
		l, _ := json.Marshal(left.Detail)
		r, _ := json.Marshal(right.Detail)
		return string(l) < string(r)
	})
	return a.problems
}

func (a *analysis) addDeclarations(section int, variables []model.Variable, path string) {
	a.scopes[section+1] = map[string][]int{}
	for i, v := range variables {
		d := declaration{variable: v, section: section, path: fmt.Sprintf("%s[%d]", path, i)}
		if section >= 0 && legalSection(a.plan.Sections[section].ID) {
			d.sectionID = a.plan.Sections[section].ID
		}
		if legalName(v.Name) && (section < 0 || d.sectionID != "") {
			scope := d.sectionID
			if section < 0 {
				scope = "plan"
			}
			d.question = scope + "." + v.Name
		}
		index := len(a.declarations)
		a.declarations = append(a.declarations, d)
		a.scopes[section+1][v.Name] = append(a.scopes[section+1][v.Name], index)
	}
}

func (a *analysis) emit(code, path string, d declaration) {
	severity := model.SeverityError
	if code == "condition-empty" {
		severity = model.SeverityWarning
	}
	a.problems = append(a.problems, model.Problem{Code: "authoring." + code, Severity: severity, Message: repairMessage(code), Question: d.question, Section: d.sectionID, Detail: map[string]string{"path": path}})
}

func (a *analysis) sectionLocation(index int) declaration {
	d := declaration{section: index}
	if legalSection(a.plan.Sections[index].ID) {
		d.sectionID = a.plan.Sections[index].ID
	}
	return d
}

func (a *analysis) checkVariable(index int) {
	d := a.declarations[index]
	v := d.variable
	if !legalName(v.Name) {
		a.emit("variable-name", d.path+".name", d)
	}
	if v.Audience != "" && v.Audience != model.AudienceOnboarder && v.Audience != model.AudienceOnboardee {
		a.emit("variable-audience", d.path+".audience", d)
	}
	if !validType(v) {
		a.emit("variable-type", d.path+".type", d)
	} else {
		if v.Multiple && variableType(v) != model.VariableTypeContact && variableType(v) != model.VariableTypeRole {
			a.emit("multiple", d.path+".multiple", d)
		}
		if !validOptions(v) {
			a.emit("choice-options", d.path+".options", d)
		}
		if !validPattern(v) {
			a.emit("pattern", d.path+".pattern", d)
		}
		if !a.validValue(v, v.Default, false) {
			a.emit("default", d.path+".default", d)
		}
	}
	a.checkSecrets(v.Default, d.path+".default", d)
	if d.section < 0 && v.AskIf != nil {
		a.emit("plan-ask-if", d.path+".askIf", d)
	}
}

func validOptions(v model.Variable) bool {
	if variableType(v) != model.VariableTypeChoice {
		return true
	}
	if len(v.Options) == 0 {
		return false
	}
	for _, s := range v.Options {
		if s == "" {
			return false
		}
	}
	return true
}

func validPattern(v model.Variable) bool {
	if variableType(v) != model.VariableTypeString || v.Pattern == "" {
		return true
	}
	_, err := regexp.Compile(v.Pattern)
	return err == nil
}

func (a *analysis) validValue(v model.Variable, value model.Value, predicate bool) bool {
	if !predicate && (value.Kind == model.ValueKindNone || value.Kind == model.ValueKindDeferred) {
		return true
	}
	expected := model.ValueKindText
	if variableType(v) == model.VariableTypeBoolean {
		expected = model.ValueKindFlag
	}
	if listType(v) {
		expected = model.ValueKindList
	}
	if value.Kind != expected || (predicate && (expected == model.ValueKindList || (expected == model.ValueKindText && value.Text == ""))) {
		return false
	}
	if value.Kind == model.ValueKindFlag {
		return true
	}
	items := []string{value.Text}
	if value.Kind == model.ValueKindList {
		items = value.List
	}
	for _, item := range items {
		switch variableType(v) {
		case model.VariableTypeString:
			if v.Pattern != "" && validPattern(v) {
				pattern, _ := regexp.Compile(v.Pattern)
				if !pattern.MatchString(item) {
					return false
				}
			}
		case model.VariableTypeChoice:
			if validOptions(v) && !contains(v.Options, item) {
				return false
			}
		case model.VariableTypeContact, model.VariableTypeRole, model.VariableTypeTeam:
			if !a.directoryContains(variableType(v), item) {
				return false
			}
		}
	}
	return true
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func (a *analysis) directoryContains(kind model.VariableType, id string) bool {
	if id == "" {
		return false
	}
	switch kind {
	case model.VariableTypeContact:
		for _, entry := range a.plan.Directory.Contacts {
			if entry.ID == id {
				return true
			}
		}
	case model.VariableTypeRole:
		for _, entry := range a.plan.Directory.Roles {
			if entry.ID == id {
				return true
			}
		}
	case model.VariableTypeTeam:
		for _, entry := range a.plan.Directory.Teams {
			if entry.ID == id {
				return true
			}
		}
	}
	return false
}

func (a *analysis) checkSecrets(value model.Value, path string, d declaration) {
	for _, problem := range a.checker.Check(value) {
		detail := make(map[string]string, len(problem.Detail)+1)
		for key, value := range problem.Detail {
			detail[key] = value
		}
		detail["path"] = path
		problem.Detail = detail
		problem.Question = d.question
		problem.Section = d.sectionID
		a.problems = append(a.problems, problem)
	}
}

func (a *analysis) checkDirectory() {
	check := func(id, name, path string) {
		if !slugValid(id) {
			a.emit("directory-id", path+".id", declaration{})
		}
		if strings.TrimSpace(name) == "" {
			a.emit("directory-name", path+".name", declaration{})
		}
	}
	for i, e := range a.plan.Directory.Contacts {
		check(e.ID, e.Name, fmt.Sprintf("directory.contacts[%d]", i))
	}
	for i, e := range a.plan.Directory.Roles {
		check(e.ID, e.Name, fmt.Sprintf("directory.roles[%d]", i))
	}
	for i, e := range a.plan.Directory.Teams {
		check(e.ID, e.Name, fmt.Sprintf("directory.teams[%d]", i))
	}
}

func validTiming(value model.Timing) bool {
	switch value {
	case "", model.TimingBeforeDayOne, model.TimingWeekOne, model.TimingDays8To30, model.TimingDays31To90, model.TimingAnytime:
		return true
	}
	return false
}
