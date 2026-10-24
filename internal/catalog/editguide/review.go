package editguide

import (
	"bytes"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	editing "onboard-kit/contracts/catalog/editor"
	creation "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/model"
)

type display struct {
	bytes strings.Builder
	full  bool
}

func (text *display) add(value string) {
	if text.full {
		return
	}
	if len(value) > 8388608-text.bytes.Len() {
		text.full = true
		return
	}
	text.bytes.WriteString(value)
}

func (text *display) line(value string) { text.add(value); text.add("\n") }

func (text *display) quoted(value string) {
	text.add("\"")
	for _, character := range value {
		switch character {
		case '"':
			text.add("\\\"")
		case '\\':
			text.add("\\\\")
		case '\n':
			text.add("\\n")
		case '\r':
			text.add("\\r")
		case '\t':
			text.add("\\t")
		default:
			if unicode.IsPrint(character) && !unicode.Is(unicode.Cf, character) && character != '\u2028' && character != '\u2029' {
				text.add(string(character))
				continue
			}
			if character > 0xffff {
				first, second := utf16.EncodeRune(character)
				text.escape(first)
				text.escape(second)
			} else {
				text.escape(character)
			}
		}
		if text.full {
			return
		}
	}
	text.add("\"")
}

func (text *display) escape(value rune) {
	digits := "0123456789abcdef"
	text.add(string([]byte{'\\', 'u', digits[(value>>12)&15], digits[(value>>8)&15], digits[(value>>4)&15], digits[value&15]}))
}

func quote(value string) string { var text display; text.quoted(value); return text.bytes.String() }

func (text *display) document(label string, value []byte) {
	text.line(label + ":")
	for {
		index := bytes.IndexByte(value, '\n')
		part := value
		if index >= 0 {
			part = value[:index]
		}
		text.add("  | ")
		text.quoted(string(part))
		text.add("\n")
		if index < 0 || text.full {
			return
		}
		value = value[index+1:]
	}
}

func (text *display) problems(problems []editing.Diagnostic) {
	text.line("Problems: " + strconv.Itoa(len(problems)))
	for _, problem := range problems {
		text.add("  - Stage: ")
		text.quoted(problem.Stage)
		text.add(" Path: ")
		text.quoted(problem.Path)
		text.add(" Code: ")
		text.quoted(problem.Code)
		text.add(" Severity: ")
		text.quoted(string(problem.Severity))
		text.add(" Repair: ")
		text.quoted(problem.Message)
		text.add("\n")
	}
}

func problemText(problems []editing.Diagnostic) string {
	var text display
	text.problems(problems)
	return text.bytes.String()
}

func (text *display) payload(value model.Value) {
	switch value.Kind {
	case model.ValueKindText:
		text.add("  Text: ")
		text.quoted(value.Text)
		text.add("\n")
	case model.ValueKindFlag:
		text.line("  Flag: " + strconv.FormatBool(value.Flag))
	case model.ValueKindList:
		text.line("  List: " + strconv.Itoa(len(value.List)))
		for _, item := range value.List {
			text.add("  - ")
			text.quoted(item)
			text.add("\n")
		}
	}
}

func valueName(value model.Value) string {
	switch value.Kind {
	case model.ValueKindText:
		return "text"
	case model.ValueKindFlag:
		return "flag"
	case model.ValueKindList:
		return "list"
	case model.ValueKindDeferred:
		return "deferred"
	}
	return "none"
}

func (driver command) review(original model.Plan, raw []byte, path string, operations []editing.Operation) (string, []byte, bool, []editing.Diagnostic, bool) {
	proposed, problems := driver.editing.Edit(original, operations)
	if diagnosticErrors(problems) {
		return "", nil, false, problems, false
	}
	for _, stage := range []string{"authoring", "plan"} {
		var findings []model.Problem
		if stage == "authoring" {
			findings = driver.authoring.Validate(proposed)
		} else {
			findings = driver.checking.Validate(proposed)
		}
		batch := project(findings, stage)
		if problemErrors(findings) {
			return "", nil, false, batch, false
		}
		problems = append(problems, batch...)
	}
	encoded, diagnostics, faithful := driver.encodeFaithfully(proposed)
	if !faithful {
		return "", nil, false, diagnostics, false
	}
	unchanged := equivalent(reflect.ValueOf(original), reflect.ValueOf(proposed), false)
	if unchanged {
		encoded = raw
	}
	guides := map[string]creation.Guidance{}
	for _, operation := range operations {
		if operation.Kind != editing.Add {
			continue
		}
		bindings := make([]creation.Binding, len(operation.Bindings))
		for i, binding := range operation.Bindings {
			bindings[i] = creation.Binding{Name: binding.Name, Value: binding.Value}
		}
		result, diagnostics := driver.creation.Create(creation.Request{PlanID: original.ID, Directory: original.Directory, Instances: []creation.Instance{{SourceID: operation.SourceID, PermanentSectionID: operation.SectionID, Bindings: bindings}}})
		var batch []editing.Diagnostic
		for _, diagnostic := range diagnostics {
			batch = append(batch, project([]model.Problem{diagnostic.Problem}, "guidance")...)
		}
		if diagnosticErrors(batch) {
			return "", nil, false, batch, false
		}
		problems = append(problems, batch...)
		for _, guide := range result.Guidance {
			previous, exists := guides[guide.ID]
			if exists && !bytes.Equal(previous.Markdown, guide.Markdown) {
				return "", nil, false, nil, false
			}
			if !exists {
				previous = creation.Guidance{ID: guide.ID, Markdown: guide.Markdown}
			}
			previous.InstanceIDs = append(previous.InstanceIDs, guide.InstanceIDs...)
			guides[guide.ID] = previous
		}
	}
	var text display
	text.line("Catalog edit review — not saved. No personal work or outcomes are changed.")
	text.line("Plan: " + quote(path))
	text.line("Operations: " + strconv.Itoa(len(operations)))
	for i, operation := range operations {
		text.add("  - " + strconv.Itoa(i+1) + " Kind: ")
		text.quoted(string(operation.Kind))
		text.add(" Instance: ")
		text.quoted(operation.SectionID)
		text.add("\n")
		switch operation.Kind {
		case editing.Add:
			text.add("  Source: ")
			text.quoted(operation.SourceID)
			text.add("\n")
			text.line("  Initial bindings: " + strconv.Itoa(len(operation.Bindings)))
			for _, binding := range operation.Bindings {
				text.add("  Binding: ")
				text.quoted(binding.Name)
				text.add(" Value: ")
				text.quoted(valueName(binding.Value))
				text.add("\n")
				text.payload(binding.Value)
			}
		case editing.Submit, editing.Unset:
			text.add("  Variable: ")
			text.quoted(operation.Name)
			text.add("\n")
			if operation.Kind == editing.Submit {
				text.add("  Value: ")
				text.quoted(valueName(operation.Value))
				text.add("\n")
				text.payload(operation.Value)
			}
		}
	}
	text.document("Original plan bytes", raw)
	text.document("Proposed plan bytes", encoded)
	if unchanged {
		text.line("Save action: unchanged; original bytes retained.")
	} else {
		text.line("Save action: replace.")
	}
	text.problems(problems)
	text.line("Preparation guidance for newly added instances only: " + strconv.Itoa(len(guides)))
	ids := make([]string, 0, len(guides))
	for id := range guides {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		guide := guides[id]
		sort.Strings(guide.InstanceIDs)
		var reasons []string
		for _, reason := range guide.InstanceIDs {
			if len(reasons) == 0 || reason != reasons[len(reasons)-1] {
				reasons = append(reasons, reason)
			}
		}
		text.add("Guide: ")
		text.quoted(id)
		text.add("\n")
		text.line("Instance reasons: " + strconv.Itoa(len(reasons)))
		for _, reason := range reasons {
			text.add("  - ")
			text.quoted(reason)
			text.add("\n")
		}
		text.document("Advisory instructions", guide.Markdown)
	}
	text.line("Advice includes retired new additions; no arrangement or activity is completed.")
	text.add(reviewPrompt)
	if text.full {
		return "", nil, false, nil, false
	}
	return text.bytes.String(), encoded, unchanged, nil, true
}
