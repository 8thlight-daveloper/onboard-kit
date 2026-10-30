package editguide

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	editorcontract "onboard-kit/contracts/catalog/editor"
	materializercontract "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/model"
)

const (
	draftPrompt   = "Edit draft — not saved. Choose add, submit, unset, retire, reactivate, drop, review, or cancel:\n"
	reviewPrompt  = "Review complete. Choose save, back, or cancel:\n"
	invalidAction = "Choose one of the displayed actions.\n"
	limitReached  = "Edit limit reached; drop a staged operation before adding another.\n"
	maxReviewText = 8388608
)

const startupText = "Catalog edit draft — not saved.\n" +
	"Only this selected plan may be replaced after complete review and explicit save.\n" +
	"No record or preparation is read; saved personal answers and outcomes are not edited.\n" +
	"Existing Submit empty values clear; initial Add bindings retain exact empty values.\n" +
	"Do not enter secrets or unnecessary personal information. Scrollback may retain shared facts.\n" +
	"Input limit: 1048576 bytes total; 65536 bytes per logical line.\n" +
	"Find source IDs with catalog list and catalog show --id ID.\n" +
	"Plan: @PLAN@\n" +
	draftPrompt

type display struct {
	text strings.Builder
	full bool
}

func (d *display) add(text string) {
	if d.full {
		return
	}
	if len(text) > maxReviewText-d.text.Len() {
		d.full = true
		return
	}
	d.text.WriteString(text)
}

func (d *display) scalar(text string) {
	d.add("\"")
	for _, r := range text {
		if d.full {
			return
		}
		switch {
		case r == '"':
			d.add("\\\"")
		case r == '\\':
			d.add("\\\\")
		case r == '\n':
			d.add("\\n")
		case r == '\r':
			d.add("\\r")
		case r == '\t':
			d.add("\\t")
		case unicode.IsPrint(r) && !unicode.Is(unicode.Cf, r) && r != 0x2028 && r != 0x2029:
			d.add(string(r))
		case r <= 0xffff:
			d.add(fmt.Sprintf("\\u%04x", r))
		default:
			high, low := utf16.EncodeRune(r)
			d.add(fmt.Sprintf("\\u%04x\\u%04x", high, low))
		}
	}
	d.add("\"")
}

func quote(text string) string {
	var d display
	d.scalar(text)
	return d.text.String()
}

func (d *display) document(label string, data []byte) {
	d.add(label + ":\n")
	text := string(data)
	for {
		part, rest, found := strings.Cut(text, "\n")
		d.add("  | ")
		d.scalar(part)
		d.add("\n")
		if d.full || !found {
			return
		}
		text = rest
	}
}

type problem struct {
	stage, path, code string
	severity          model.Severity
	repair            string
}

func (d *display) problems(problems []problem) {
	d.add(fmt.Sprintf("Problems: %d\n", len(problems)))
	for _, p := range problems {
		d.add("  - Stage: ")
		d.scalar(p.stage)
		d.add(" Path: ")
		d.scalar(p.path)
		d.add(" Code: ")
		d.scalar(p.code)
		d.add(" Severity: ")
		d.scalar(string(p.severity))
		d.add(" Repair: ")
		d.scalar(p.repair)
		d.add("\n")
	}
}

func problemText(problems []problem) string {
	var d display
	d.problems(problems)
	return d.text.String()
}

var stageRepairs = map[string]string{
	"decode":    "Repair the selected plan's format; duplicate keys and invalid spellings must be fixed explicitly.",
	"encode":    "Preserve or explicitly repair unrepresentable plan data before saving.",
	"authoring": "Repair the selected plan's declarations before editing.",
	"plan":      "Repair the selected plan's structure or stored facts before editing.",
	"guidance":  "Ask the catalog maintainer to repair the advisory preparation source.",
}

var directoryRepairs = map[string]string{
	"directory.contact-unreachable": "Supply an approved business route for the declared contact.",
	"directory.role-vacant":         "Name an accountable holder for the declared role.",
	"directory.email-malformed":     "Repair the declared business email address.",
}

func dependencyProblem(stage, code string, severity model.Severity) problem {
	repair := stageRepairs[stage]
	if override, found := directoryRepairs[code]; found {
		repair = override
	}
	if strings.HasPrefix(code, "secret.") {
		repair = "Remove credential-like content before saving."
	}
	return problem{stage: stage, path: "plan", code: code, severity: severity, repair: repair}
}

func modelProblems(stage string, problems []model.Problem) []problem {
	projected := make([]problem, len(problems))
	for i, p := range problems {
		projected[i] = dependencyProblem(stage, p.Code, p.Severity)
	}
	return projected
}

func guidanceProblems(diagnostics []materializercontract.Diagnostic) []problem {
	projected := make([]problem, len(diagnostics))
	for i, diagnostic := range diagnostics {
		projected[i] = dependencyProblem("guidance", diagnostic.Problem.Code, diagnostic.Problem.Severity)
	}
	return projected
}

func editorProblems(diagnostics []editorcontract.Diagnostic) []problem {
	projected := make([]problem, len(diagnostics))
	for i, diagnostic := range diagnostics {
		projected[i] = problem{stage: diagnostic.Stage, path: diagnostic.Path, code: diagnostic.Code, severity: diagnostic.Severity, repair: diagnostic.Message}
	}
	return projected
}

func hasErrors(problems []problem) bool {
	for _, p := range problems {
		if p.severity == model.SeverityError {
			return true
		}
	}
	return false
}

func hasModelErrors(problems []model.Problem) bool {
	for _, p := range problems {
		if p.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func kindName(kind model.ValueKind) string {
	switch kind {
	case model.ValueKindText:
		return "text"
	case model.ValueKindFlag:
		return "flag"
	case model.ValueKindList:
		return "list"
	case model.ValueKindDeferred:
		return "deferred"
	}
	return ""
}

func (d *display) payload(value model.Value) {
	switch value.Kind {
	case model.ValueKindText:
		d.add("  Text: ")
		d.scalar(value.Text)
		d.add("\n")
	case model.ValueKindFlag:
		d.add("  Flag: " + strconv.FormatBool(value.Flag) + "\n")
	case model.ValueKindList:
		d.add(fmt.Sprintf("  List: %d\n", len(value.List)))
		for _, item := range value.List {
			d.add("  - ")
			d.scalar(item)
			d.add("\n")
		}
	}
}

func (d *display) namedValue(label, name string, value model.Value) {
	d.add("  " + label + ": ")
	d.scalar(name)
	d.add(" Value: ")
	d.scalar(kindName(value.Kind))
	d.add("\n")
	d.payload(value)
}

func (d *display) operations(operations []editorcontract.Operation) {
	d.add(fmt.Sprintf("Operations: %d\n", len(operations)))
	for i, operation := range operations {
		d.add(fmt.Sprintf("  - %d Kind: ", i+1))
		d.scalar(string(operation.Kind))
		d.add(" Instance: ")
		d.scalar(operation.SectionID)
		d.add("\n")
		switch operation.Kind {
		case editorcontract.Add:
			d.add("  Source: ")
			d.scalar(operation.SourceID)
			d.add(fmt.Sprintf("\n  Initial bindings: %d\n", len(operation.Bindings)))
			for _, binding := range operation.Bindings {
				d.namedValue("Binding", binding.Name, binding.Value)
			}
		case editorcontract.Submit:
			d.add("  Variable: ")
			d.scalar(operation.Name)
			d.add("\n  Value: ")
			d.scalar(kindName(operation.Value.Kind))
			d.add("\n")
			d.payload(operation.Value)
		case editorcontract.Unset:
			d.add("  Variable: ")
			d.scalar(operation.Name)
			d.add("\n")
		}
	}
}
