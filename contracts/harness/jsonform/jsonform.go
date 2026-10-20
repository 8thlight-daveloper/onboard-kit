package jsonform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"onboard-kit/internal/model"
)

func Plan(data []byte) (model.Plan, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return model.Plan{}, err
	}
	r := &reader{}
	plan := r.plan("plan", raw)
	return plan, r.err
}

func Section(data []byte) (model.Section, error) {
	raw, err := decodeJSON(data)
	if err != nil {
		return model.Section{}, err
	}
	r := &reader{}
	section := r.section("section", raw)
	return section, r.err
}

func Condition(data []byte) (model.Condition, error) {
	raw, err := decodeJSON(data)
	if err != nil {
		return model.Condition{}, err
	}
	r := &reader{}
	condition := r.condition("condition", raw)
	if condition == nil {
		return model.Condition{}, r.err
	}
	return *condition, r.err
}

func Value(data []byte) (model.Value, error) {
	raw, err := decodeJSON(data)
	if err != nil {
		return model.Value{}, err
	}
	r := &reader{}
	value := r.value("value", raw)
	return value, r.err
}

func Answers(data []byte) (model.Answers, error) {
	raw, err := decodeJSON(data)
	if err != nil {
		return model.Answers{}, err
	}
	r := &reader{}
	answers := r.answers("answers", raw)
	return answers, r.err
}

func decodeJSON(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

type reader struct {
	err error
}

func (r *reader) fail(path, format string, args ...any) {
	if r.err == nil {
		r.err = fmt.Errorf("%s: %s", path, fmt.Sprintf(format, args...))
	}
}

func (r *reader) object(path string, v any, allowed ...string) map[string]any {
	if v == nil {
		return nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		r.fail(path, "want an object, got %T", v)
		return nil
	}
	allow := map[string]bool{}
	for _, key := range allowed {
		allow[key] = true
	}
	for key := range m {
		if !allow[key] {
			r.fail(path, "unknown key %q", key)
		}
	}
	return m
}

func (r *reader) str(path string, m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		r.fail(path+"."+key, "want a string, got %T", v)
	}
	return s
}

func (r *reader) boolean(path string, m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		r.fail(path+"."+key, "want a boolean, got %T", v)
	}
	return b
}

func (r *reader) integer(path string, m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	n, ok := v.(json.Number)
	if !ok {
		r.fail(path+"."+key, "want an integer, got %T", v)
		return 0
	}
	i, err := strconv.Atoi(n.String())
	if err != nil {
		r.fail(path+"."+key, "want an integer, got %s", n)
	}
	return i
}

func (r *reader) list(path string, m map[string]any, key string) []any {
	v, ok := m[key]
	if !ok {
		return nil
	}
	l, ok := v.([]any)
	if !ok {
		r.fail(path+"."+key, "want a list, got %T", v)
	}
	return l
}

func (r *reader) strings(path string, m map[string]any, key string) []string {
	var out []string
	for i, item := range r.list(path, m, key) {
		s, ok := item.(string)
		if !ok {
			r.fail(fmt.Sprintf("%s.%s[%d]", path, key, i), "want a string, got %T", item)
		}
		out = append(out, s)
	}
	return out
}

func (r *reader) value(path string, v any) model.Value {
	switch x := v.(type) {
	case nil:
		return model.Value{}
	case string:
		return model.Value{Kind: model.ValueKindText, Text: x}
	case bool:
		return model.Value{Kind: model.ValueKindFlag, Flag: x}
	case []any:
		list := []string{}
		for i, item := range x {
			s, ok := item.(string)
			if !ok {
				r.fail(fmt.Sprintf("%s[%d]", path, i), "want a string, got %T", item)
			}
			list = append(list, s)
		}
		return model.Value{Kind: model.ValueKindList, List: list}
	case map[string]any:
		m := r.object(path, x, "deferred")
		if b, ok := m["deferred"].(bool); ok && b {
			return model.Value{Kind: model.ValueKindDeferred}
		}
		r.fail(path, `want {"deferred": true}`)
	default:
		r.fail(path, "want a string, boolean, list, or deferred marker, got %T", v)
	}
	return model.Value{}
}

func (r *reader) valueAt(path string, m map[string]any, key string) model.Value {
	v, ok := m[key]
	if !ok {
		return model.Value{}
	}
	return r.value(path+"."+key, v)
}

func (r *reader) plan(path string, v any) model.Plan {
	m := r.object(path, v, "schemaVersion", "provenance", "variables", "answers", "sections", "unknownFields")
	plan := model.Plan{
		SchemaVersion: r.integer(path, m, "schemaVersion"),
		Provenance:    r.provenance(path+".provenance", m["provenance"]),
		Answers:       r.answers(path+".answers", m["answers"]),
		UnknownFields: r.unknownFields(path+".unknownFields", m["unknownFields"]),
	}
	for i, item := range r.list(path, m, "variables") {
		plan.Variables = append(plan.Variables, r.variable(fmt.Sprintf("%s.variables[%d]", path, i), item))
	}
	for i, item := range r.list(path, m, "sections") {
		plan.Sections = append(plan.Sections, r.section(fmt.Sprintf("%s.sections[%d]", path, i), item))
	}
	return plan
}

func (r *reader) provenance(path string, v any) model.Provenance {
	m := r.object(path, v, "templateSet", "generatedOn")
	return model.Provenance{
		TemplateSet: r.str(path, m, "templateSet"),
		GeneratedOn: model.Date(r.str(path, m, "generatedOn")),
	}
}

func (r *reader) answers(path string, v any) model.Answers {
	m := r.object(path, v, "plan", "sections")
	answers := model.Answers{}
	if plan := r.object(path+".plan", m["plan"], allKeys(m["plan"])...); plan != nil {
		answers.Plan = map[string]model.Value{}
		for name, raw := range plan {
			answers.Plan[name] = r.value(path+".plan."+name, raw)
		}
	}
	if sections := r.object(path+".sections", m["sections"], allKeys(m["sections"])...); sections != nil {
		answers.Sections = map[string]map[string]model.Value{}
		for id, raw := range sections {
			inner := r.object(path+".sections."+id, raw, allKeys(raw)...)
			answers.Sections[id] = map[string]model.Value{}
			for name, value := range inner {
				answers.Sections[id][name] = r.value(path+".sections."+id+"."+name, value)
			}
		}
	}
	return answers
}

func allKeys(v any) []string {
	m, _ := v.(map[string]any)
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func (r *reader) variable(path string, v any) model.Variable {
	m := r.object(path, v, "name", "description", "type", "multiple", "options", "pattern", "patternHint", "required", "default", "audience", "askIf")
	variable := model.Variable{
		Name:        r.str(path, m, "name"),
		Description: r.str(path, m, "description"),
		Type:        model.VariableType(r.str(path, m, "type")),
		Multiple:    r.boolean(path, m, "multiple"),
		Options:     r.strings(path, m, "options"),
		Pattern:     r.str(path, m, "pattern"),
		PatternHint: r.str(path, m, "patternHint"),
		Required:    r.boolean(path, m, "required"),
		Default:     r.valueAt(path, m, "default"),
		Audience:    model.Audience(r.str(path, m, "audience")),
	}
	if raw, ok := m["askIf"]; ok {
		variable.AskIf = r.condition(path+".askIf", raw)
	}
	return variable
}

func (r *reader) condition(path string, v any) *model.Condition {
	m := r.object(path, v, "all")
	condition := &model.Condition{}
	for i, item := range r.list(path, m, "all") {
		condition.All = append(condition.All, r.clause(fmt.Sprintf("%s.all[%d]", path, i), item))
	}
	return condition
}

func (r *reader) clause(path string, v any) model.Clause {
	m := r.object(path, v, "question", "answer", "section", "enabled")
	_, hasSection := m["section"]
	_, hasEnabled := m["enabled"]
	if hasSection || hasEnabled {
		return model.Clause{Section: r.str(path, m, "section"), Enabled: r.boolean(path, m, "enabled")}
	}
	test := r.object(path+".answer", m["answer"], "is", "isNot", "isOneOf", "isNotOneOf", "isSettled")
	clause := model.Clause{Question: r.str(path, m, "question")}
	clause.Answer.Is = r.valueAt(path+".answer", test, "is")
	clause.Answer.IsNot = r.valueAt(path+".answer", test, "isNot")
	for i, item := range r.list(path+".answer", test, "isOneOf") {
		clause.Answer.IsOneOf = append(clause.Answer.IsOneOf, r.value(fmt.Sprintf("%s.answer.isOneOf[%d]", path, i), item))
	}
	for i, item := range r.list(path+".answer", test, "isNotOneOf") {
		clause.Answer.IsNotOneOf = append(clause.Answer.IsNotOneOf, r.value(fmt.Sprintf("%s.answer.isNotOneOf[%d]", path, i), item))
	}
	if _, ok := test["isSettled"]; ok {
		b := r.boolean(path+".answer", test, "isSettled")
		clause.Answer.IsSettled = &b
	}
	return clause
}

func (r *reader) section(path string, v any) model.Section {
	m := r.object(path, v, "id", "title", "summary", "includeIf", "timing", "enabled", "after", "owner", "escalation", "verification", "body", "variables", "steps", "source", "unknownFields")
	section := model.Section{
		ID:            r.str(path, m, "id"),
		Title:         r.str(path, m, "title"),
		Summary:       r.str(path, m, "summary"),
		Timing:        model.Timing(r.str(path, m, "timing")),
		After:         r.strings(path, m, "after"),
		Owner:         r.str(path, m, "owner"),
		Escalation:    r.strings(path, m, "escalation"),
		Verification:  r.str(path, m, "verification"),
		Body:          r.str(path, m, "body"),
		UnknownFields: r.unknownFields(path+".unknownFields", m["unknownFields"]),
	}
	if raw, ok := m["includeIf"]; ok {
		section.IncludeIf = r.condition(path+".includeIf", raw)
	}
	if _, ok := m["enabled"]; ok {
		enabled := r.boolean(path, m, "enabled")
		section.Enabled = &enabled
	}
	for i, item := range r.list(path, m, "variables") {
		section.Variables = append(section.Variables, r.variable(fmt.Sprintf("%s.variables[%d]", path, i), item))
	}
	for i, item := range r.list(path, m, "steps") {
		section.Steps = append(section.Steps, r.step(fmt.Sprintf("%s.steps[%d]", path, i), item))
	}
	if raw, ok := m["source"]; ok {
		s := r.object(path+".source", raw, "template", "hash")
		section.Source = &model.Source{Template: r.str(path+".source", s, "template"), Hash: r.str(path+".source", s, "hash")}
	}
	return section
}

func (r *reader) step(path string, v any) model.Step {
	m := r.object(path, v, "id", "name", "notes", "owner", "timing", "escalation", "run", "open", "directory", "env", "checks")
	step := model.Step{
		ID:         r.str(path, m, "id"),
		Name:       r.str(path, m, "name"),
		Notes:      r.str(path, m, "notes"),
		Owner:      r.str(path, m, "owner"),
		Timing:     model.Timing(r.str(path, m, "timing")),
		Escalation: r.strings(path, m, "escalation"),
		Run:        r.strings(path, m, "run"),
		Open:       r.str(path, m, "open"),
		Directory:  r.str(path, m, "directory"),
	}
	if env := r.object(path+".env", m["env"], allKeys(m["env"])...); env != nil {
		step.Env = map[string]string{}
		for name, raw := range env {
			s, ok := raw.(string)
			if !ok {
				r.fail(path+".env."+name, "want a string, got %T", raw)
			}
			step.Env[name] = s
		}
	}
	for i, item := range r.list(path, m, "checks") {
		cpath := fmt.Sprintf("%s.checks[%d]", path, i)
		c := r.object(cpath, item, "run", "ask", "env")
		step.Checks = append(step.Checks, model.Check{
			Run: r.strings(cpath, c, "run"),
			Ask: r.str(cpath, c, "ask"),
			Env: r.str(cpath, c, "env"),
		})
	}
	return step
}

func (r *reader) unknownFields(path string, v any) map[string]any {
	if v == nil {
		return nil
	}
	m := r.object(path, v, allKeys(v)...)
	out := map[string]any{}
	for key, raw := range m {
		out[key] = r.generic(path+"."+key, raw)
	}
	return out
}

func (r *reader) generic(path string, v any) any {
	switch x := v.(type) {
	case nil, string, bool:
		return x
	case json.Number:
		if !strings.ContainsAny(x.String(), ".eE") {
			i, err := strconv.ParseInt(x.String(), 10, 64)
			if err != nil {
				r.fail(path, "bad integer %s", x)
			}
			return i
		}
		f, err := x.Float64()
		if err != nil {
			r.fail(path, "bad number %s", x)
		}
		return f
	case []any:
		out := make([]any, 0, len(x))
		for i, item := range x {
			out = append(out, r.generic(fmt.Sprintf("%s[%d]", path, i), item))
		}
		return out
	case map[string]any:
		out := map[string]any{}
		for key, item := range x {
			out[key] = r.generic(path+"."+key, item)
		}
		return out
	}
	r.fail(path, "unsupported kind %T", v)
	return nil
}
