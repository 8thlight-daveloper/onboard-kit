package jsonform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"onboard-kit/internal/model"
)

type pair struct {
	key   string
	value any
}

type object []pair

func (o *object) add(key string, value any) {
	*o = append(*o, pair{key, value})
}

func (o *object) addText(key, value string) {
	if value != "" {
		o.add(key, value)
	}
}

func (o *object) addFlag(key string, value bool) {
	if value {
		o.add(key, true)
	}
}

func (o *object) addStrings(key string, values []string) {
	if len(values) > 0 {
		o.add(key, stringsOf(values))
	}
}

func stringsOf(values []string) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

func JSON(plan model.Plan) ([]byte, error) {
	var out bytes.Buffer
	if err := write(&out, planObject(plan), 0); err != nil {
		return nil, err
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

func SectionJSON(section model.Section) ([]byte, error) {
	var out bytes.Buffer
	if err := write(&out, sectionObject(section), 0); err != nil {
		return nil, err
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

func planObject(plan model.Plan) object {
	var o object
	if plan.SchemaVersion != 0 {
		o.add("schemaVersion", plan.SchemaVersion)
	}
	o.addText("id", plan.ID)
	if plan.Provenance != (model.Provenance{}) {
		var p object
		p.addText("templateSet", plan.Provenance.TemplateSet)
		p.addText("generatedOn", string(plan.Provenance.GeneratedOn))
		o.add("provenance", p)
	}
	if plan.Origin != (model.Origin{}) {
		var origin object
		origin.addText("planId", plan.Origin.PlanID)
		origin.addText("source", plan.Origin.Source)
		origin.addText("contentHash", plan.Origin.ContentHash)
		origin.addText("refreshedOn", string(plan.Origin.RefreshedOn))
		o.add("origin", origin)
	}
	if directory := directoryObject(plan.Directory); len(directory) > 0 {
		o.add("directory", directory)
	}
	if len(plan.Variables) > 0 {
		o.add("variables", mapList(plan.Variables, variableObject))
	}
	if plan.Answers.Plan != nil || plan.Answers.Sections != nil {
		o.add("answers", answersObject(plan.Answers))
	}
	if len(plan.Sections) > 0 {
		o.add("sections", mapList(plan.Sections, sectionObject))
	}
	if plan.UnknownFields != nil {
		o.add("unknownFields", plan.UnknownFields)
	}
	return o
}

func mapList[T any](items []T, convert func(T) object) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, convert(item))
	}
	return out
}

func answersObject(answers model.Answers) object {
	var o object
	if answers.Plan != nil {
		o.add("plan", valueMap(answers.Plan))
	}
	if answers.Sections != nil {
		var sections object
		for _, id := range sortedKeys(answers.Sections) {
			sections.add(id, valueMap(answers.Sections[id]))
		}
		o.add("sections", sections)
	}
	return o
}

func valueMap(values map[string]model.Value) object {
	o := object{}
	for _, name := range sortedKeys(values) {
		o.add(name, valueOf(values[name]))
	}
	return o
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func valueOf(value model.Value) any {
	switch value.Kind {
	case model.ValueKindText:
		return value.Text
	case model.ValueKindFlag:
		return value.Flag
	case model.ValueKindList:
		return stringsOf(value.List)
	case model.ValueKindDeferred:
		return object{{"deferred", true}}
	}
	return nil
}

func variableObject(variable model.Variable) object {
	var o object
	o.addText("name", variable.Name)
	o.addText("description", variable.Description)
	o.addText("example", variable.Example)
	o.addText("type", string(variable.Type))
	o.addFlag("multiple", variable.Multiple)
	o.addStrings("options", variable.Options)
	o.addText("pattern", variable.Pattern)
	o.addText("patternHint", variable.PatternHint)
	o.addFlag("required", variable.Required)
	if variable.Default.Kind != model.ValueKindNone {
		o.add("default", valueOf(variable.Default))
	}
	o.addText("audience", string(variable.Audience))
	if variable.AskIf != nil {
		o.add("askIf", conditionObject(*variable.AskIf))
	}
	return o
}

func conditionObject(condition model.Condition) object {
	var o object
	if len(condition.All) > 0 {
		o.add("all", mapList(condition.All, clauseObject))
	}
	return o
}

func clauseObject(clause model.Clause) object {
	var o object
	if clause.Section != "" || (clause.Enabled && clause.Question == "") {
		o.addText("section", clause.Section)
		o.add("enabled", clause.Enabled)
		return o
	}
	o.addText("question", clause.Question)
	var test object
	if clause.Answer.Is.Kind != model.ValueKindNone {
		test.add("is", valueOf(clause.Answer.Is))
	}
	if clause.Answer.IsNot.Kind != model.ValueKindNone {
		test.add("isNot", valueOf(clause.Answer.IsNot))
	}
	if len(clause.Answer.IsOneOf) > 0 {
		test.add("isOneOf", mapValues(clause.Answer.IsOneOf))
	}
	if len(clause.Answer.IsNotOneOf) > 0 {
		test.add("isNotOneOf", mapValues(clause.Answer.IsNotOneOf))
	}
	if clause.Answer.IsSettled != nil {
		test.add("isSettled", *clause.Answer.IsSettled)
	}
	o.add("answer", test)
	return o
}

func mapValues(values []model.Value) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, valueOf(value))
	}
	return out
}

func sectionObject(section model.Section) object {
	var o object
	o.addText("id", section.ID)
	o.addText("title", section.Title)
	o.addText("summary", section.Summary)
	o.addText("parent", section.Parent)
	if section.IncludeIf != nil {
		o.add("includeIf", conditionObject(*section.IncludeIf))
	}
	o.addText("timing", string(section.Timing))
	if section.Enabled != nil {
		o.add("enabled", *section.Enabled)
	}
	o.addStrings("after", section.After)
	o.addText("owner", section.Owner)
	o.addStrings("escalation", section.Escalation)
	o.addText("verification", section.Verification)
	o.addText("body", section.Body)
	if len(section.Variables) > 0 {
		o.add("variables", mapList(section.Variables, variableObject))
	}
	if len(section.Steps) > 0 {
		o.add("steps", mapList(section.Steps, stepObject))
	}
	if section.Source != nil {
		var source object
		source.addText("template", section.Source.Template)
		source.addText("hash", section.Source.Hash)
		o.add("source", source)
	}
	if section.UnknownFields != nil {
		o.add("unknownFields", section.UnknownFields)
	}
	return o
}

func stepObject(step model.Step) object {
	var o object
	o.addText("id", step.ID)
	o.addText("name", step.Name)
	o.addText("notes", step.Notes)
	o.addText("owner", step.Owner)
	o.addText("timing", string(step.Timing))
	o.addStrings("escalation", step.Escalation)
	o.addStrings("run", step.Run)
	o.addText("open", step.Open)
	o.addText("directory", step.Directory)
	if step.Env != nil {
		env := object{}
		for _, name := range sortedKeys(step.Env) {
			env.add(name, step.Env[name])
		}
		o.add("env", env)
	}
	if len(step.Checks) > 0 {
		o.add("checks", mapList(step.Checks, checkObject))
	}
	return o
}

func checkObject(check model.Check) object {
	var o object
	o.addStrings("run", check.Run)
	o.addText("ask", check.Ask)
	o.addText("env", check.Env)
	return o
}

func write(out *bytes.Buffer, value any, depth int) error {
	switch x := value.(type) {
	case nil:
		out.WriteString("null")
	case bool:
		out.WriteString(strconv.FormatBool(x))
	case int:
		out.WriteString(strconv.Itoa(x))
	case int64:
		out.WriteString(strconv.FormatInt(x, 10))
	case float64:
		out.WriteString(decimalText(x))
	case string:
		return writeText(out, x)
	case object:
		return writeObject(out, x, depth)
	case []any:
		return writeList(out, x, depth)
	case map[string]any:
		var o object
		for _, key := range sortedKeys(x) {
			o.add(key, x[key])
		}
		return writeObject(out, o, depth)
	default:
		return fmt.Errorf("cannot write %T as JSON", value)
	}
	return nil
}

func decimalText(f float64) string {
	text := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(text, ".") {
		text += ".0"
	}
	return text
}

func writeText(out *bytes.Buffer, text string) error {
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(text); err != nil {
		return err
	}
	out.Truncate(out.Len() - 1)
	return nil
}

func writeObject(out *bytes.Buffer, o object, depth int) error {
	if len(o) == 0 {
		out.WriteString("{}")
		return nil
	}
	out.WriteString("{\n")
	for i, p := range o {
		out.WriteString(strings.Repeat("  ", depth+1))
		if err := writeText(out, p.key); err != nil {
			return err
		}
		out.WriteString(": ")
		if err := write(out, p.value, depth+1); err != nil {
			return err
		}
		if i < len(o)-1 {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	out.WriteString(strings.Repeat("  ", depth) + "}")
	return nil
}

func writeList(out *bytes.Buffer, items []any, depth int) error {
	if len(items) == 0 {
		out.WriteString("[]")
		return nil
	}
	out.WriteString("[\n")
	for i, item := range items {
		out.WriteString(strings.Repeat("  ", depth+1))
		if err := write(out, item, depth+1); err != nil {
			return err
		}
		if i < len(items)-1 {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	out.WriteString(strings.Repeat("  ", depth) + "]")
	return nil
}

func directoryObject(directory model.Directory) object {
	var o object
	if len(directory.Contacts) > 0 {
		o.add("contacts", mapList(directory.Contacts, func(contact model.Contact) object {
			var c object
			c.addText("id", contact.ID)
			c.addText("name", contact.Name)
			c.addText("email", contact.Email)
			c.addText("slack", contact.Slack)
			c.addText("slackId", contact.SlackID)
			return c
		}))
	}
	if len(directory.Roles) > 0 {
		o.add("roles", mapList(directory.Roles, func(role model.Role) object {
			var r object
			r.addText("id", role.ID)
			r.addText("name", role.Name)
			r.addStrings("holders", role.Holders)
			return r
		}))
	}
	if len(directory.Teams) > 0 {
		o.add("teams", mapList(directory.Teams, func(team model.Team) object {
			var t object
			t.addText("id", team.ID)
			t.addText("name", team.Name)
			t.addText("contact", team.Contact)
			if team.Holders != nil {
				var held object
				for _, role := range sortedKeys(team.Holders) {
					held.add(role, stringsOf(team.Holders[role]))
				}
				t.add("holders", held)
			}
			return t
		}))
	}
	return o
}
