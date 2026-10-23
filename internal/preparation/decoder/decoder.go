package decoder

import (
	"bytes"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

type decoder struct{}

func New() decoder { return decoder{} }

type diagnostic struct {
	problem                      model.Problem
	line, column, rank, sequence int
}

type reader struct{ diagnostics []diagnostic }

func issue(code, path string) model.Problem {
	detail := map[string]string(nil)
	if path != "" {
		detail = map[string]string{"path": path}
	}
	message := "Correct the preparation field to match its declared type."
	switch code {
	case "file.not-yaml":
		message = "Use valid UTF-8 YAML and correct its syntax."
	case "file.not-a-preparation":
		message = "Supply a preparation YAML mapping document."
	case "file.multiple-documents":
		message = "Keep exactly one preparation document in the file."
	case "version.invalid", "version.too-old", "version.too-new":
		message = "Use the supported schemaVersion: 1 as an unquoted integer."
	case "format.boolean-spelling":
		message = "Use true or false for booleans, or quote the value as text."
	case "format.number-spelling":
		message = "Quote numeric text; use canonical decimal numbers for unknown numeric data."
	case "format.null-spelling":
		message = "Use null for an empty value, or quote the value as text."
	case "format.alias":
		message = "Remove anchors and expand aliases or merge keys into explicit values."
	case "format.duplicate-key":
		message = "Remove the repeated mapping key; each key must be unique."
	case "format.unknown-field":
		message = "Remove the unknown field or correct its exact field name."
	case "format.empty-item":
		message = "Supply a value for the list item or remove the empty item."
	}
	if strings.HasPrefix(code, "kind.") {
		message = "Expected kind: preparation."
	}
	return model.Problem{Code: code, Severity: model.SeverityError, Message: message, Detail: detail}
}

func failure(code, path string) (model.Preparation, []model.Problem) {
	return model.Preparation{}, []model.Problem{issue(code, path)}
}

func (decoder) Decode(data []byte) (model.Preparation, []model.Problem) {
	if !utf8.Valid(data) {
		return failure("file.not-yaml", "")
	}
	parser := yaml.NewDecoder(bytes.NewReader(data))
	var documents []*yaml.Node
	for {
		var document yaml.Node
		err := parser.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			return failure("file.not-yaml", "")
		}
		documents = append(documents, &document)
	}
	if len(documents) == 0 || len(documents[0].Content) == 0 {
		return failure("file.not-a-preparation", "")
	}
	root := documents[0].Content[0]
	if root.Kind == yaml.ScalarNode && !quoted(root) && root.Value == "" && root.Anchor == "" {
		return failure("file.not-a-preparation", "")
	}
	if len(documents) > 1 {
		return failure("file.multiple-documents", "")
	}
	if root.Kind != yaml.MappingNode {
		return failure("file.not-a-preparation", "")
	}
	version, count := field(root, "schemaVersion")
	if count > 1 {
		return failure("format.duplicate-key", "schemaVersion")
	}
	if version == nil || version.Kind != yaml.ScalarNode || quoted(version) || !integer.MatchString(version.Value) {
		return failure("version.invalid", "schemaVersion")
	}
	value, err := strconv.ParseInt(version.Value, 10, 64)
	if err != nil {
		return failure("version.invalid", "schemaVersion")
	}
	if value < 1 {
		return failure("version.too-old", "schemaVersion")
	}
	if value > 1 {
		return failure("version.too-new", "schemaVersion")
	}
	kind, count := field(root, "kind")
	if count > 1 {
		return failure("format.duplicate-key", "kind")
	}
	if kind == nil || kind.Kind == yaml.ScalarNode && !quoted(kind) && (kind.Value == "" || kind.Value == "null") {
		return failure("kind.missing", "kind")
	}
	if kind.Kind != yaml.ScalarNode || kind.Value != "preparation" {
		return failure("kind.not-a-preparation", "kind")
	}
	r := reader{}
	result := model.Preparation{SchemaVersion: 1, Kind: "preparation"}
	r.mapping(root, "", root, func(key string, node, place *yaml.Node, path string) {
		switch key {
		case "schemaVersion", "kind":
			r.anchor(node, path, place)
		case "account":
			result.Account = r.text(node, path, place, false)
		case "planId":
			result.PlanID = r.text(node, path, place, false)
		case "recipient":
			result.Recipient = r.recipient(node, path, place)
		case "arrangements":
			result.Arrangements = r.arrangements(node, path, place)
		default:
			if result.UnknownFields == nil {
				result.UnknownFields = map[string]any{}
			}
			result.UnknownFields[key] = r.unknown(node, path, place)
		}
	})
	if len(r.diagnostics) == 0 {
		return result, nil
	}
	sort.SliceStable(r.diagnostics, func(i, j int) bool {
		a, b := r.diagnostics[i], r.diagnostics[j]
		if a.line != b.line {
			return a.line < b.line
		}
		if a.column != b.column {
			return a.column < b.column
		}
		if a.rank != b.rank {
			return a.rank < b.rank
		}
		return a.sequence < b.sequence
	})
	problems := make([]model.Problem, len(r.diagnostics))
	for i, diagnostic := range r.diagnostics {
		problems[i] = diagnostic.problem
	}
	return model.Preparation{}, problems
}

func field(node *yaml.Node, key string) (*yaml.Node, int) {
	var found *yaml.Node
	count := 0
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Kind == yaml.ScalarNode && node.Content[i].Value == key {
			count++
			found = node.Content[i+1]
		}
	}
	return found, count
}

func quoted(node *yaml.Node) bool {
	return node.Style&(yaml.SingleQuotedStyle|yaml.DoubleQuotedStyle|yaml.LiteralStyle|yaml.FoldedStyle) != 0
}

var integer = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
var decimal = regexp.MustCompile(`^-?(0|[1-9][0-9]*)\.[0-9]+$`)
var numeric = regexp.MustCompile(`(?i)^[+-]?(?:[0-9](?:_?[0-9])*(?:\.[0-9](?:_?[0-9])*|\.)?(?:e[+-]?[0-9]+)?|\.[0-9](?:_?[0-9])*(?:e[+-]?[0-9]+)?|0x[0-9a-f](?:_?[0-9a-f])*|0o[0-7](?:_?[0-7])*|0b[01](?:_?[01])*|\.inf|\.nan)$`)

func scalar(node *yaml.Node) (any, string) {
	if quoted(node) {
		return node.Value, ""
	}
	value := node.Value
	if value == "true" {
		return true, ""
	}
	if value == "false" {
		return false, ""
	}
	switch strings.ToLower(value) {
	case "true", "false", "yes", "no", "on", "off", "y", "n":
		return nil, "format.boolean-spelling"
	}
	if value == "" || value == "null" {
		return nil, ""
	}
	if value == "~" || value == "Null" || value == "NULL" {
		return nil, "format.null-spelling"
	}
	if integer.MatchString(value) {
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, "format.number-spelling"
		}
		return number, ""
	}
	if decimal.MatchString(value) {
		number, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, "format.number-spelling"
		}
		return number, ""
	}
	if numeric.MatchString(value) {
		return nil, "format.number-spelling"
	}
	return value, ""
}

func (r *reader) add(code, path string, place *yaml.Node) {
	rank := 0
	switch code {
	case "format.boolean-spelling":
		rank = 0
	case "format.number-spelling":
		rank = 1
	case "format.null-spelling":
		rank = 2
	case "format.alias":
		rank = -1
	case "format.duplicate-key":
		rank = 4
	case "format.wrong-type":
		rank = 5
	case "format.unknown-field":
		rank = 6
	case "format.empty-item":
		rank = 7
	}
	r.diagnostics = append(r.diagnostics, diagnostic{issue(code, path), place.Line, place.Column, rank, len(r.diagnostics)})
}

func (r *reader) anchor(node *yaml.Node, path string, place *yaml.Node) bool {
	if node.Anchor != "" || node.Kind == yaml.AliasNode {
		r.add("format.alias", path, place)
	}
	return node.Kind == yaml.AliasNode
}

func join(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

func (r *reader) mapping(node *yaml.Node, path string, place *yaml.Node, read func(string, *yaml.Node, *yaml.Node, string)) {
	if r.anchor(node, path, place) {
		return
	}
	if node.Kind == yaml.ScalarNode {
		value, code := scalar(node)
		if code != "" {
			r.add(code, path, place)
			return
		}
		if value == nil {
			return
		}
	}
	if node.Kind != yaml.MappingNode {
		r.add("format.wrong-type", path, place)
		return
	}
	seen := map[string]bool{}
	for i := 0; i < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Kind != yaml.ScalarNode {
			r.add("format.wrong-type", path, place)
			continue
		}
		child := join(path, key.Value)
		if key.Value == "<<" && !quoted(key) {
			r.add("format.alias", child, key)
			continue
		}
		r.anchor(key, child, key)
		if seen[key.Value] {
			r.add("format.duplicate-key", child, key)
		}
		seen[key.Value] = true
		read(key.Value, value, key, child)
	}
}

func (r *reader) text(node *yaml.Node, path string, place *yaml.Node, item bool) string {
	if r.anchor(node, path, place) {
		return ""
	}
	if node.Kind != yaml.ScalarNode {
		r.add("format.wrong-type", path, place)
		return ""
	}
	value, code := scalar(node)
	if code != "" {
		r.add(code, path, place)
		return ""
	}
	if value == nil {
		if item {
			r.add("format.empty-item", path, place)
		}
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	switch value.(type) {
	case int64, float64:
		r.add("format.number-spelling", path, place)
	default:
		r.add("format.wrong-type", path, place)
	}
	return ""
}

func (r *reader) sequence(node *yaml.Node, path string, place *yaml.Node, read func(*yaml.Node, string)) {
	if r.anchor(node, path, place) {
		return
	}
	if node.Kind == yaml.ScalarNode {
		value, code := scalar(node)
		if code != "" {
			r.add(code, path, place)
			return
		}
		if value == nil {
			return
		}
	}
	if node.Kind != yaml.SequenceNode {
		r.add("format.wrong-type", path, place)
		return
	}
	for i, item := range node.Content {
		read(item, path+"["+strconv.Itoa(i)+"]")
	}
}

func (r *reader) recipient(node *yaml.Node, path string, place *yaml.Node) model.PreparationRecipient {
	result := model.PreparationRecipient{}
	r.mapping(node, path, place, func(key string, value, keyNode *yaml.Node, child string) {
		var target *string
		switch key {
		case "name":
			target = &result.Name
		case "companyEmail":
			target = &result.CompanyEmail
		case "plannedStartDate":
			result.PlannedStartDate = model.Date(r.text(value, child, keyNode, false))
			return
		case "team":
			target = &result.Team
		case "manager":
			target = &result.Manager
		case "onboardingOwner":
			target = &result.OnboardingOwner
		case "buddy":
			target = &result.Buddy
		case "workLocation":
			target = &result.WorkLocation
		case "timeZone":
			target = &result.TimeZone
		default:
			r.add("format.unknown-field", child, keyNode)
			return
		}
		*target = r.text(value, child, keyNode, false)
	})
	return result
}

func (r *reader) arrangements(node *yaml.Node, path string, place *yaml.Node) []model.Arrangement {
	var result []model.Arrangement
	r.sequence(node, path, place, func(item *yaml.Node, child string) {
		if item.Kind == yaml.ScalarNode {
			value, code := scalar(item)
			if value == nil && code == "" {
				r.anchor(item, child, item)
				r.add("format.empty-item", child, item)
				return
			}
		}
		arrangement := model.Arrangement{}
		r.mapping(item, child, item, func(key string, value, keyNode *yaml.Node, fieldPath string) {
			var target *string
			switch key {
			case "id":
				target = &arrangement.ID
			case "section":
				target = &arrangement.Section
			case "kind":
				target = &arrangement.Kind
			case "owner":
				target = &arrangement.Owner
			case "reference":
				target = &arrangement.Reference
			case "status":
				target = &arrangement.Status
			case "note":
				target = &arrangement.Note
			case "targetDate":
				arrangement.TargetDate = model.Date(r.text(value, fieldPath, keyNode, false))
				return
			case "contacts":
				r.sequence(value, fieldPath, keyNode, func(contact *yaml.Node, contactPath string) {
					arrangement.Contacts = append(arrangement.Contacts, r.text(contact, contactPath, contact, true))
				})
				return
			default:
				r.add("format.unknown-field", fieldPath, keyNode)
				return
			}
			*target = r.text(value, fieldPath, keyNode, false)
		})
		result = append(result, arrangement)
	})
	return result
}

func (r *reader) unknown(node *yaml.Node, path string, place *yaml.Node) any {
	if node.Kind == yaml.MappingNode {
		result := map[string]any{}
		r.mapping(node, path, place, func(key string, value, keyNode *yaml.Node, child string) {
			result[key] = r.unknown(value, child, keyNode)
		})
		return result
	}
	if r.anchor(node, path, place) {
		return nil
	}
	if node.Kind == yaml.SequenceNode {
		result := make([]any, len(node.Content))
		for i, item := range node.Content {
			result[i] = r.unknown(item, path+"["+strconv.Itoa(i)+"]", item)
		}
		return result
	}
	value, code := scalar(node)
	if code != "" {
		r.add(code, path, place)
	}
	return value
}
