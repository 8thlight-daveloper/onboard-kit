package creator

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf16"

	materializercontract "onboard-kit/contracts/catalog/materializer"
)

type display struct {
	text strings.Builder
	full bool
}

func (d *display) add(text string) {
	if d.full {
		return
	}
	if len(text) > 8388608-d.text.Len() {
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
		switch r {
		case '"':
			d.add("\\\"")
		case '\\':
			d.add("\\\\")
		case '\n':
			d.add("\\n")
		case '\r':
			d.add("\\r")
		case '\t':
			d.add("\\t")
		default:
			if unicode.IsPrint(r) && !unicode.Is(unicode.Cf, r) && r != 0x2028 && r != 0x2029 {
				d.add(string(r))
			} else if r <= 0xffff {
				d.add(fmt.Sprintf("\\u%04x", r))
			} else {
				high, low := utf16.EncodeRune(r)
				d.add(fmt.Sprintf("\\u%04x\\u%04x", high, low))
			}
		}
	}
	d.add("\"")
}

func quote(text string) string { var d display; d.scalar(text); return d.text.String() }

func (d *display) document(label string, bytes []byte) {
	d.add(label + ":\n")
	text := string(bytes)
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

func problemText(problems []materializercontract.Diagnostic) string {
	var d display
	d.problems(problems)
	return d.text.String()
}

func (d *display) problems(problems []materializercontract.Diagnostic) {
	d.add(fmt.Sprintf("Problems: %d\n", len(problems)))
	for _, problem := range problems {
		d.add("  - Stage: ")
		d.scalar(problem.Stage)
		d.add(" Path: ")
		d.scalar(problem.Path)
		d.add(" Code: ")
		d.scalar(problem.Problem.Code)
		d.add(" Severity: ")
		d.scalar(string(problem.Problem.Severity))
		d.add(" Repair: ")
		d.scalar(repair(problem.Stage, problem.Problem.Code))
		d.add("\n")
	}
}

func repair(stage, code string) string {
	switch code {
	case "catalog.unknown-id":
		return "Choose an exact section-template ID from catalog list."
	case "materialize.carrier":
		return "Select a section-template source for an activity."
	case "materialize.instance-id":
		return "Supply a nonreserved lowercase permanent section ID."
	case "materialize.duplicate-instance":
		return "Give each independent activity its own permanent section ID."
	case "materialize.duplicate-binding":
		return "Supply each scoped binding name once."
	case "materialize.binding-name":
		return "Bind an exact variable declared in this scope."
	case "materialize.binding-owner":
		return "Keep newcomer answers outside the shared organization plan."
	case "materialize.limit":
		return "Reduce the selection or held creation data to the supported limits."
	case "materialize.text":
		return "Use valid UTF8 creation data."
	case "materialize.value-kind":
		return "Choose a supported value kind."
	case "materialize.plan-id":
		return "Supply an explicit nonempty plan identity."
	case "answer.wrong-kind":
		return "Use a value kind that matches the declared variable type."
	case "answer.not-an-option":
		return "Use an exact declared choice option."
	case "answer.pattern-mismatch":
		return "Use a value matching the declared pattern."
	case "directory.unknown-id":
		return "Keep this binding unset until its directory target is declared."
	}
	if strings.HasPrefix(code, "secret.") {
		return "Remove credential-like content; supply a nonsecret organization fact."
	}
	switch stage {
	case "request", "binding":
		return "Correct the indicated instance or scoped binding before review."
	case "source":
		return "Choose an exact section-template ID from catalog list."
	case "load", "profile", "authoring":
		return "Ask the catalog maintainer to repair this definition."
	case "structure", "plan":
		return "Repair the plan definition or scoped facts before review."
	case "answer":
		return "Use a value accepted by this variable, or leave the binding unset."
	case "guidance":
		return "Ask the catalog maintainer to repair the preparation guide."
	case "encode":
		return "Remove or replace the unrepresentable draft value before review."
	default:
		return "Resolve the indicated creation problem before saving."
	}
}

func (s *session) review() ([]byte, bool) {
	result, problems := s.command.materializer.Create(s.request)
	if hasErrors(problems) {
		s.write("Review unavailable.\n" + problemText(problems))
		return nil, false
	}
	bytes, encodingProblems := s.command.encoder.Encode(result.Plan)
	if len(encodingProblems) > 0 {
		batch := make([]materializercontract.Diagnostic, len(encodingProblems))
		for i, problem := range encodingProblems {
			batch[i] = materializercontract.Diagnostic{Stage: "encode", Problem: problem}
		}
		s.write("Review unavailable.\n" + problemText(batch))
		return nil, false
	}
	if len(bytes) > 1048576 {
		s.write("Review unavailable: supported output limits exceeded.\n")
		return nil, false
	}
	var d display
	d.add("Catalog review — not saved. Saving does not establish invitation readiness.\nPlan: ")
	d.scalar(s.request.PlanID)
	d.add("\nDestination: ")
	d.scalar(s.parent.path)
	d.add("\n")
	d.problems(problems)
	d.document("Plan file", bytes)
	d.add(fmt.Sprintf("Preparation guidance: %d\n", len(result.Guidance)))
	for _, guide := range result.Guidance {
		d.add("Guide: ")
		d.scalar(guide.ID)
		d.add(fmt.Sprintf("\nInstance reasons: %d\n", len(guide.InstanceIDs)))
		for _, id := range guide.InstanceIDs {
			d.add("  - ")
			d.scalar(id)
			d.add("\n")
		}
		d.document("Advisory instructions", guide.Markdown)
	}
	d.add("Preparation guidance is advisory; no personal arrangement or activity is completed.\n" + reviewPrompt)
	if d.full {
		s.write("Review unavailable: supported output limits exceeded.\n")
		return nil, false
	}
	if !s.write(d.text.String()) {
		return nil, false
	}
	return bytes, true
}
