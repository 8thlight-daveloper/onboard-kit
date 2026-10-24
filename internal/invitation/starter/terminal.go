package starter

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	reviewapi "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

const displayLimit = 8388608

type textBuilder struct {
	strings.Builder
	failed bool
	limit  int
}

func (b *textBuilder) add(s string) {
	if b.failed || len(s) > b.limit-b.Len() {
		b.failed = true
		return
	}
	b.WriteString(s)
}
func (b *textBuilder) line(s string)             { b.add(s); b.add("\n") }
func (b *textBuilder) field(label, value string) { b.line(label + ": " + quote(value)) }
func (b *textBuilder) count(label string, n int) { b.line(fmt.Sprintf("%s: %d", label, n)) }
func (b *textBuilder) flag(label string, v bool) { b.line(fmt.Sprintf("%s: %t", label, v)) }
func (b *textBuilder) document(label, s string) {
	b.line(label + ":")
	for _, part := range strings.Split(s, "\n") {
		b.line("  | " + quote(part))
	}
}
func (b *textBuilder) list(label string, values []string) {
	b.count(label, len(values))
	for _, s := range values {
		b.line("  - " + quote(s))
	}
}
func (b *textBuilder) value(label string, v model.Value) {
	switch v.Kind {
	case model.ValueKindNone:
		b.line(label + ": none")
	case model.ValueKindDeferred:
		b.line(label + ": deferred")
	case model.ValueKindText:
		b.line(label + ": text " + quote(v.Text))
	case model.ValueKindFlag:
		b.line(fmt.Sprintf("%s: flag %t", label, v.Flag))
	case model.ValueKindList:
		b.line(fmt.Sprintf("%s: list %d", label, len(v.List)))
		for _, s := range v.List {
			b.line("  - " + quote(s))
		}
	default:
		b.failed = true
	}
}
func (b *textBuilder) diagnostics(label string, items []reviewapi.Diagnostic) {
	b.count(label, len(items))
	for _, d := range items {
		severity := string(d.Problem.Severity)
		if severity != "error" && severity != "warning" {
			severity = "unknown"
		}
		b.line("  - Severity: " + quote(severity) + " Repair: " + quote(d.Display))
	}
}
func quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			if !unicode.IsPrint(r) || unicode.Is(unicode.Cf, r) || r == 0x2028 || r == 0x2029 {
				if r <= 0xffff {
					fmt.Fprintf(&b, "\\u%04x", r)
				} else {
					a, c := utf16.EncodeRune(r)
					fmt.Fprintf(&b, "\\u%04x\\u%04x", a, c)
				}
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
func freeze(review reviewapi.Review) (string, error) {
	b := textBuilder{limit: displayLimit}
	b.line("Read-only onboarding session.")
	b.line("Instructions are displayed, not executed. Dates are goals, not completion evidence.")
	b.line("Defaults and suggestions are not accepted answers.")
	b.count("Activities", len(review.Session.Activities))
	for i, a := range review.Session.Activities {
		s := a.Section
		b.count("Activity", i+1)
		b.field("ID", s.ID)
		b.field("Title", s.Title)
		b.field("Summary", s.Summary)
		b.field("Parent", s.Parent)
		b.field("Owner", s.Owner)
		b.field("Timing", string(s.Timing))
		b.field("Verification", s.Verification)
		b.field("Current definition hash", a.CurrentHash)
		b.document("Body", s.Body)
		b.list("Requires", s.Requires)
		b.list("After", s.After)
		b.list("Waiting on", a.WaitingOn)
		b.list("Escalation", a.Escalation)
		b.count("Steps", len(s.Steps))
		for j, step := range s.Steps {
			b.count("Step", j+1)
			b.field("ID", step.ID)
			b.field("Name", step.Name)
			b.field("Owner", step.Owner)
			b.field("Timing", string(step.Timing))
			b.field("Open", step.Open)
			b.field("Directory", step.Directory)
			b.document("Notes", step.Notes)
			b.list("Run", step.Run)
			b.list("Escalation", step.Escalation)
			keys := make([]string, 0, len(step.Env))
			for k := range step.Env {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			b.count("Environment", len(keys))
			for _, k := range keys {
				b.line("  - Key: " + quote(k) + " Value: " + quote(step.Env[k]))
			}
			b.count("Checks", len(step.Checks))
			for k, check := range step.Checks {
				b.count("Check", k+1)
				b.list("Run", check.Run)
				b.document("Ask", check.Ask)
				b.field("Environment question", check.Env)
			}
		}
	}
	b.count("Questions", len(review.Session.Questions))
	for i, q := range review.Session.Questions {
		b.count("Question", i+1)
		b.field("ID", q.ID)
		b.field("Scope", q.Scope)
		b.field("Name", q.Name)
		b.field("Description", q.Description)
		b.field("Example", q.Example)
		b.field("Type", string(q.Type))
		b.field("Audience", string(q.Audience))
		b.flag("Required", q.Required)
		b.flag("Multiple", q.Multiple)
		b.value("Answer", q.Answer)
		b.value("Default suggestion", q.DeclaredDefault)
		b.list("Options", q.Options)
		b.count("Choices", len(q.Choices))
		for _, c := range q.Choices {
			b.line("  - Value: " + quote(c.Value) + " Label: " + quote(c.Label))
		}
		b.count("Suggestions", len(q.Suggestions))
		for j, s := range q.Suggestions {
			b.count("Suggestion", j+1)
			b.field("Source", string(s.Source))
			b.value("Value", s.Value)
		}
	}
	blockers := make([]string, len(review.Session.Blockers))
	for i, b := range review.Session.Blockers {
		blockers[i] = b.Question
	}
	b.list("Blockers", blockers)
	b.count("Recorded outcomes", len(review.Session.Outcomes))
	for i, h := range review.Session.Outcomes {
		b.count("Recorded outcome", i+1)
		b.field("Section", h.Outcome.Section)
		b.field("Outcome", string(h.Outcome.Kind))
		b.field("On", string(h.Outcome.On))
		b.field("Recorded definition hash", h.Outcome.SectionHash)
		b.field("Current definition hash", h.CurrentHash)
		b.field("Comparison", h.Comparison)
		b.document("Note", h.Outcome.Note)
		b.document("Differently", h.Outcome.Differently)
	}
	b.flag("Support available", review.Support.Available)
	if !review.Support.Available {
		b.line("Current support is unavailable; ask the onboarding owner to repair preparation.")
		b.diagnostics("Support problems", review.PreparationProblems)
	} else {
		r := review.Support.Recipient
		b.field("Name", r.Name)
		b.field("Company email", r.CompanyEmail)
		b.field("Planned start date (goal)", string(r.PlannedStartDate))
		b.field("Team ID", r.Team)
		b.field("Manager contact ID", r.Manager)
		b.field("Onboarding owner contact ID", r.OnboardingOwner)
		b.field("Buddy contact ID", r.Buddy)
		b.field("Work location", r.WorkLocation)
		b.field("Time zone", r.TimeZone)
		b.count("Arrangements", len(review.Support.Arrangements))
		for i, a := range review.Support.Arrangements {
			b.count("Arrangement", i+1)
			b.field("ID", a.ID)
			b.field("Section", a.Section)
			b.field("Kind", a.Kind)
			b.field("Owner contact ID", a.Owner)
			b.field("Reference", a.Reference)
			b.field("Target date (goal)", string(a.TargetDate))
			b.field("Status", a.Status)
			b.list("Helper contact IDs", a.Contacts)
			b.document("Note", a.Note)
		}
	}
	b.diagnostics("Session problems", review.SessionProblems)
	b.document("Invitation-time guidance", string(review.InvitationFirstSteps))
	if b.failed {
		return "", errLimit
	}
	return b.String(), nil
}

type lineInput struct {
	reader   io.Reader
	consumed int
}

func (input *lineInput) line() (string, bool, error) {
	var data []byte
	var one [1]byte
	for {
		if input.consumed >= 1048576 {
			return "", false, errInputLimit
		}
		n, err := input.reader.Read(one[:])
		if n < 0 || n > 1 {
			return "", false, errInput
		}
		if n == 1 {
			input.consumed++
			if one[0] == '\n' {
				if len(data) > 0 && data[len(data)-1] == '\r' {
					data = data[:len(data)-1]
				}
				if len(data) > 65536 || !utf8.Valid(data) {
					return "", false, errInput
				}
				if err != nil && err != io.EOF {
					return "", false, errInput
				}
				return string(data), true, nil
			}
			data = append(data, one[0])
			if len(data) > 65537 || (len(data) == 65537 && data[len(data)-1] != '\r') {
				return "", false, errInput
			}
		}
		if err != nil {
			if err != io.EOF {
				return "", false, errInput
			}
			if !utf8.Valid(data) || len(data) > 65536 {
				return "", false, errInput
			}
			return "", false, nil
		}
		if n == 0 {
			return "", false, errInput
		}
	}
}
func writeAll(w io.Writer, s string) error {
	n, err := w.Write([]byte(s))
	if err != nil || n != len(s) {
		return errOutput
	}
	return nil
}
func choose(input *lineInput, w io.Writer, prompt string, actions ...string) (string, error) {
	for {
		if err := writeAll(w, prompt); err != nil {
			return "", err
		}
		line, complete, err := input.line()
		if err != nil {
			return "", err
		}
		if !complete {
			return "", nil
		}
		for _, a := range actions {
			if line == a {
				return a, nil
			}
		}
		if err := writeAll(w, "Choose one of the displayed actions.\n"); err != nil {
			return "", err
		}
	}
}
