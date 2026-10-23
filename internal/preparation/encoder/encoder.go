package encoder

import (
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	contract "onboard-kit/contracts/preparation/encoder"
	"onboard-kit/internal/model"
)

type encoder struct{}

type entry struct {
	key   string
	value node
}

type node struct {
	scalar string
	block  string
	fields []entry
	items  []node
	kind   byte
}

type traversal struct {
	problems []model.Problem
	branch   map[container]bool
}

type container struct {
	kind    reflect.Kind
	pointer uintptr
	length  int
}

func New() contract.Encoder { return encoder{} }

func (encoder) Encode(preparation model.Preparation) ([]byte, []model.Problem) {
	t := traversal{branch: make(map[container]bool)}
	root := node{kind: 'm'}
	root.fields = append(root.fields, entry{"schemaVersion", node{scalar: strconv.Itoa(preparation.SchemaVersion)}}, entry{"kind", t.text(preparation.Kind, "kind")})
	t.field(&root, "account", preparation.Account, "account")
	t.field(&root, "planId", preparation.PlanID, "planId")
	recipient := node{kind: 'm'}
	values := []entryText{
		{"name", preparation.Recipient.Name}, {"companyEmail", preparation.Recipient.CompanyEmail},
		{"plannedStartDate", string(preparation.Recipient.PlannedStartDate)}, {"team", preparation.Recipient.Team},
		{"manager", preparation.Recipient.Manager}, {"onboardingOwner", preparation.Recipient.OnboardingOwner},
		{"buddy", preparation.Recipient.Buddy}, {"workLocation", preparation.Recipient.WorkLocation}, {"timeZone", preparation.Recipient.TimeZone},
	}
	for _, value := range values {
		t.field(&recipient, value.key, value.value, "recipient."+value.key)
	}
	if len(recipient.fields) > 0 {
		root.fields = append(root.fields, entry{"recipient", recipient})
	}
	arrangements := node{kind: 'l'}
	for index, arrangement := range preparation.Arrangements {
		path := "arrangements[" + strconv.Itoa(index) + "]"
		item := node{kind: 'm'}
		for _, value := range []entryText{{"id", arrangement.ID}, {"section", arrangement.Section}, {"kind", arrangement.Kind}, {"owner", arrangement.Owner}} {
			t.field(&item, value.key, value.value, path+"."+value.key)
		}
		if len(arrangement.Contacts) > 0 {
			contacts := node{kind: 'l'}
			for i, contact := range arrangement.Contacts {
				contacts.items = append(contacts.items, t.text(contact, path+".contacts["+strconv.Itoa(i)+"]"))
			}
			item.fields = append(item.fields, entry{"contacts", contacts})
		}
		for _, value := range []entryText{{"reference", arrangement.Reference}, {"targetDate", string(arrangement.TargetDate)}, {"status", arrangement.Status}, {"note", arrangement.Note}} {
			t.field(&item, value.key, value.value, path+"."+value.key)
		}
		if len(item.fields) == 0 {
			t.problem("encode.empty-item", path)
		}
		arrangements.items = append(arrangements.items, item)
	}
	if len(arrangements.items) > 0 {
		root.fields = append(root.fields, entry{"arrangements", arrangements})
	}
	for _, key := range keys(preparation.UnknownFields) {
		if slices.Contains([]string{"schemaVersion", "kind", "account", "planId", "recipient", "arrangements"}, key) {
			t.problem("encode.unknown-field-collision", key)
			continue
		}
		if !utf8.ValidString(key) {
			t.problem("encode.unsupported-value", key)
			continue
		}
		root.fields = append(root.fields, entry{key, t.unknown(preparation.UnknownFields[key], key)})
	}
	if len(t.problems) > 0 {
		return nil, t.problems
	}
	var output strings.Builder
	emit(&output, root, "", 0)
	return []byte(output.String()), nil
}

type entryText struct{ key, value string }

func (t *traversal) field(mapping *node, key, value, path string) {
	if value != "" {
		mapping.fields = append(mapping.fields, entry{key, t.text(value, path)})
	}
}

func (t *traversal) problem(code, path string) {
	t.problems = append(t.problems, model.Problem{Code: code, Severity: model.SeverityError, Message: "Cannot encode this field.", Detail: map[string]string{"path": path}})
}

func (t *traversal) text(value, path string) node {
	if !utf8.ValidString(value) {
		t.problem("encode.unsupported-value", path)
		return node{}
	}
	if literal(value) {
		return node{kind: 'b', block: value}
	}
	return node{scalar: quoted(value)}
}

func (t *traversal) unknown(value any, path string) node {
	switch value := value.(type) {
	case nil:
		return node{scalar: "null"}
	case string:
		return t.text(value, path)
	case bool:
		return node{scalar: strconv.FormatBool(value)}
	case int:
		return node{scalar: strconv.Itoa(value)}
	case int64:
		return node{scalar: strconv.FormatInt(value, 10)}
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.problem("encode.unsupported-value", path)
			return node{}
		}
		text := strconv.FormatFloat(value, 'f', -1, 64)
		if !strings.Contains(text, ".") {
			text += ".0"
		}
		return node{scalar: text}
	case map[string]any:
		identity := container{kind: reflect.Map, pointer: reflect.ValueOf(value).Pointer()}
		if t.branch[identity] {
			t.problem("encode.unsupported-value", path)
			return node{}
		}
		t.branch[identity] = true
		defer delete(t.branch, identity)
		result := node{kind: 'm'}
		for _, key := range keys(value) {
			childPath := path + "." + key
			if !utf8.ValidString(key) {
				t.problem("encode.unsupported-value", childPath)
				continue
			}
			result.fields = append(result.fields, entry{key, t.unknown(value[key], childPath)})
		}
		return result
	case []any:
		identity := container{kind: reflect.Slice, pointer: reflect.ValueOf(value).Pointer(), length: len(value)}
		if len(value) > 0 && t.branch[identity] {
			t.problem("encode.unsupported-value", path)
			return node{}
		}
		t.branch[identity] = true
		defer delete(t.branch, identity)
		result := node{kind: 'l'}
		for i, item := range value {
			result.items = append(result.items, t.unknown(item, path+"["+strconv.Itoa(i)+"]"))
		}
		return result
	default:
		t.problem("encode.unsupported-value", path)
		return node{}
	}
}

func keys(mapping map[string]any) []string {
	result := make([]string, 0, len(mapping))
	for key := range mapping {
		result = append(result, key)
	}
	slices.Sort(result)
	return result
}

func literal(value string) bool {
	if !strings.Contains(value, "\n") || value[0] == '\n' || value[0] == ' ' || strings.HasSuffix(value, "\n\n") {
		return false
	}
	for _, r := range value {
		if r == '\u2028' || r == '\u2029' || r != '\n' && unicode.IsControl(r) {
			return false
		}
	}
	for _, line := range strings.Split(value, "\n") {
		if strings.TrimRightFunc(line, unicode.IsSpace) != line {
			return false
		}
	}
	return true
}

func quoted(value string) string {
	plain := value != "" && value[0] != ' ' && !strings.HasSuffix(value, " ")
	if plain {
		first := value[0]
		plain = first >= 'a' && first <= 'z' || first >= 'A' && first <= 'Z' || first == '$' || first == '/' || first == '<'
	}
	for _, r := range value {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			plain = false
		}
	}
	if strings.Contains(value, ": ") || strings.Contains(value, " #") || strings.HasSuffix(value, ":") || value == "deferred" {
		plain = false
	}
	if slices.Contains([]string{"y", "n", "yes", "no", "true", "false", "on", "off", "null", "inf", "nan"}, strings.ToLower(value)) {
		plain = false
	}
	if plain {
		return value
	}
	var output strings.Builder
	output.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\\', '"':
			output.WriteByte('\\')
			output.WriteRune(r)
		case '\n':
			output.WriteString("\\n")
		case '\t':
			output.WriteString("\\t")
		case '\r':
			output.WriteString("\\r")
		case '\u2028':
			output.WriteString("\\u2028")
		case '\u2029':
			output.WriteString("\\u2029")
		default:
			if unicode.IsControl(r) {
				output.WriteString("\\x")
				number := strconv.FormatInt(int64(r), 16)
				if len(number) < 2 {
					output.WriteByte('0')
				}
				output.WriteString(number)
			} else {
				output.WriteRune(r)
			}
		}
	}
	output.WriteByte('"')
	return output.String()
}

func emit(output *strings.Builder, value node, prefix string, indent int) {
	switch value.kind {
	case 'm':
		if len(value.fields) == 0 {
			output.WriteString(prefix + "{}\n")
			return
		}
		for index, field := range value.fields {
			start := strings.Repeat(" ", indent)
			if index == 0 {
				start = prefix
			}
			start += quoted(field.key) + ":"
			emitField(output, field.value, start, indent)
		}
	case 'l':
		if len(value.items) == 0 {
			output.WriteString(prefix + "[]\n")
			return
		}
		for index, item := range value.items {
			start := strings.Repeat(" ", indent)
			if index == 0 {
				start = prefix
			}
			emit(output, item, start+"- ", indent+2)
		}
	case 'b':
		marker := "|-"
		if strings.HasSuffix(value.block, "\n") {
			marker = "|"
		}
		output.WriteString(prefix + marker + "\n")
		for _, line := range strings.Split(strings.TrimSuffix(value.block, "\n"), "\n") {
			if line != "" {
				output.WriteString(strings.Repeat(" ", indent) + line)
			}
			output.WriteByte('\n')
		}
	default:
		output.WriteString(prefix + value.scalar + "\n")
	}
}

func emitField(output *strings.Builder, value node, prefix string, indent int) {
	if value.kind == 'm' && len(value.fields) > 0 || value.kind == 'l' && len(value.items) > 0 {
		output.WriteString(prefix + "\n")
		emit(output, value, strings.Repeat(" ", indent+2), indent+2)
		return
	}
	emit(output, value, prefix+" ", indent+2)
}
