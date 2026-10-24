package authoringvalidator

import (
	"fmt"
	"sort"
	"strings"

	"onboard-kit/internal/model"
)

func (a *analysis) checkSection(index int) {
	s := a.plan.Sections[index]
	d := a.sectionLocation(index)
	path := fmt.Sprintf("sections[%d]", index)
	if !legalSection(s.ID) {
		a.emit("section-id", path+".id", d)
	}
	if strings.TrimSpace(s.Title) == "" {
		a.emit("required-prose", path+".title", d)
	}
	if strings.TrimSpace(s.Summary) == "" {
		a.emit("required-prose", path+".summary", d)
	}
	if !validTiming(s.Timing) {
		a.emit("timing", path+".timing", d)
	}
	for _, field := range [][2]string{{"title", s.Title}, {"summary", s.Summary}, {"owner", s.Owner}, {"verification", s.Verification}, {"body", s.Body}} {
		a.checkText(index, field[1], path+"."+field[0], field[0] == "owner", d)
	}
	a.checkEscalation(index, s.Escalation, path+".escalation", d)
	for j, step := range s.Steps {
		stepPath := fmt.Sprintf("%s.steps[%d]", path, j)
		if step.ID != "" && !slugValid(step.ID) {
			a.emit("step-id", stepPath+".id", d)
		}
		if !validTiming(step.Timing) {
			a.emit("timing", stepPath+".timing", d)
		}
		for _, field := range [][2]string{{"name", step.Name}, {"notes", step.Notes}, {"owner", step.Owner}, {"open", step.Open}, {"directory", step.Directory}} {
			a.checkText(index, field[1], stepPath+"."+field[0], field[0] == "owner", d)
		}
		for k, text := range step.Run {
			a.checkText(index, text, fmt.Sprintf("%s.run[%d]", stepPath, k), false, d)
		}
		keys := make([]string, 0, len(step.Env))
		for key := range step.Env {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for k, key := range keys {
			envPath := fmt.Sprintf("%s.env[%d]", stepPath, k)
			if !environmentValid(key) {
				a.emit("environment-name", envPath, d)
			}
			value := step.Env[key]
			a.checkText(index, value, envPath, false, d)
			a.checkSecrets(model.Value{Kind: model.ValueKindText, Text: value}, envPath, d)
		}
		a.checkEscalation(index, step.Escalation, stepPath+".escalation", d)
		for k, check := range step.Checks {
			checkPath := fmt.Sprintf("%s.checks[%d]", stepPath, k)
			operations := 0
			if len(check.Run) > 0 {
				operations++
			}
			if strings.TrimSpace(check.Ask) != "" {
				operations++
			}
			if check.Env != "" {
				operations++
			}
			if operations != 1 {
				a.emit("check", checkPath, d)
			}
			if check.Env != "" && !environmentValid(check.Env) {
				a.emit("environment-name", checkPath+".env", d)
			}
			for n, text := range check.Run {
				a.checkText(index, text, fmt.Sprintf("%s.run[%d]", checkPath, n), false, d)
			}
			a.checkText(index, check.Ask, checkPath+".ask", false, d)
		}
	}
}

func nameCharacter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= '0' && value <= '9' || value == '_'
}

func placeholders(text string) ([]string, bool) {
	var names []string
	bad := false
	for i := 0; i < len(text); {
		if text[i] != '$' {
			i++
			continue
		}
		i++
		if i < len(text) && text[i] == '$' {
			i++
			continue
		}
		if i == len(text) {
			bad = true
			continue
		}
		start := i
		if text[i] >= '0' && text[i] <= '9' {
			for i < len(text) && nameCharacter(text[i]) {
				i++
			}
			continue
		}
		if text[i] < 'a' || text[i] > 'z' {
			bad = true
			continue
		}
		for i < len(text) && nameCharacter(text[i]) {
			i++
		}
		if i-start >= 3 {
			names = append(names, text[start:i])
		}
	}
	return names, bad
}

func (a *analysis) checkText(section int, text, path string, owner bool, d declaration) {
	names, bad := placeholders(text)
	badOwner := false
	for _, name := range names {
		indices := a.scopes[section+1][name]
		if len(indices) == 0 {
			bad = true
			continue
		}
		if len(indices) > 1 {
			continue
		}
		index := indices[0]
		a.used[index] = true
		v := a.declarations[index].variable
		if owner && validType(v) && variableType(v) != model.VariableTypeContact && variableType(v) != model.VariableTypeRole {
			badOwner = true
		}
	}
	if bad {
		a.emit("dollar", path, d)
	}
	if badOwner {
		a.emit("owner-type", path, d)
	}
}

func (a *analysis) checkEscalation(section int, values []string, path string, d declaration) {
	for i, text := range values {
		bad := true
		if len(text) > 1 && text[0] == '$' && nameValid(text[1:]) {
			indices := a.scopes[section+1][text[1:]]
			if len(indices) > 1 {
				bad = false
			} else if len(indices) == 1 {
				index := indices[0]
				a.used[index] = true
				v := a.declarations[index].variable
				if !validType(v) {
					bad = false
				} else if variableType(v) == model.VariableTypeContact || variableType(v) == model.VariableTypeRole {
					bad = false
				}
			}
		}
		if bad {
			a.emit("escalation", fmt.Sprintf("%s[%d]", path, i), d)
		}
	}
}
