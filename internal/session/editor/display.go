package editor

import (
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	reviewcontract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

type builder struct {
	strings.Builder
	failed bool
}

func (b *builder) add(s string) {
	if b.failed {
		return
	}
	if len(s) > 8388608-b.Len() {
		b.failed = true
		return
	}
	_, _ = b.WriteString(s)
}
func (b *builder) line(s string)                    { b.add(s); b.add("\n") }
func (b *builder) scalar(label, s string)           { b.add(label); b.add(": "); b.quoted(s); b.add("\n") }
func (b *builder) count(label string, n int)        { b.line(label + ": " + strconv.Itoa(n)) }
func (b *builder) boolean(label string, value bool) { b.line(label + ": " + strconv.FormatBool(value)) }
func (b *builder) quoted(s string) {
	b.add("\"")
	for _, c := range s {
		switch c {
		case '"':
			b.add("\\\"")
		case '\\':
			b.add("\\\\")
		case '\n':
			b.add("\\n")
		case '\r':
			b.add("\\r")
		case '\t':
			b.add("\\t")
		default:
			if unicode.IsControl(c) || unicode.Is(unicode.Cf, c) || c == 0x2028 || c == 0x2029 || !unicode.IsPrint(c) {
				if c > 0xffff {
					hi, lo := utf16.EncodeRune(c)
					b.escape(hi)
					b.escape(lo)
				} else {
					b.escape(c)
				}
			} else {
				b.add(string(c))
			}
		}
		if b.failed {
			break
		}
	}
	b.add("\"")
}
func (b *builder) escape(c rune) {
	const digits = "0123456789abcdef"
	data := [6]byte{'\\', 'u', digits[(c>>12)&15], digits[(c>>8)&15], digits[(c>>4)&15], digits[c&15]}
	b.add(string(data[:]))
}
func quote(s string) string { b := builder{}; b.quoted(s); return b.String() }
func (b *builder) document(label, s string) {
	b.line(label + ":")
	for {
		part, rest, more := strings.Cut(s, "\n")
		b.add("  | ")
		b.quoted(part)
		b.add("\n")
		if !more || b.failed {
			break
		}
		s = rest
	}
}
func (b *builder) list(label string, items []string) {
	b.count(label, len(items))
	for _, item := range items {
		b.add("  - ")
		b.quoted(item)
		b.add("\n")
		if b.failed {
			break
		}
	}
}
func (b *builder) value(label string, value model.Value) {
	switch value.Kind {
	case model.ValueKindNone:
		b.line(label + ": none")
	case model.ValueKindDeferred:
		b.line(label + ": deferred")
	case model.ValueKindFlag:
		b.line(label + ": flag " + strconv.FormatBool(value.Flag))
	case model.ValueKindText:
		b.add(label + ": text ")
		b.quoted(value.Text)
		b.add("\n")
	case model.ValueKindList:
		b.line(label + ": list " + strconv.Itoa(len(value.List)))
		for _, item := range value.List {
			b.add("  - ")
			b.quoted(item)
			b.add("\n")
		}
	default:
		b.failed = true
	}
}
func (b *builder) diagnostics(label string, problems []reviewcontract.Diagnostic) {
	b.count(label, len(problems))
	for _, problem := range problems {
		b.add("  - Stage: ")
		b.quoted(problem.Stage)
		b.add(" Code: ")
		b.quoted(problem.Problem.Code)
		b.add(" Severity: ")
		severity := "unknown"
		if problem.Problem.Severity == model.SeverityError || problem.Problem.Severity == model.SeverityWarning {
			severity = string(problem.Problem.Severity)
		}
		b.quoted(severity)
		b.add("\n")
	}
	if label == "Session problems" {
		b.line("Review the selected source files to resolve any reported errors.")
	}
}

func identity(c *context) builder {
	b := builder{}
	b.line("Record-backed session editing is explicit.")
	b.line("Plan and preparation are organization-owned; answers and outcomes are newcomer-owned.")
	b.line("Records, terminal scrollback and private staging can contain personal data.")
	b.scalar("Registered account", c.selected.account)
	b.scalar("Registered plan", c.selected.planID)
	b.scalar("Registered root", c.selected.root)
	b.scalar("State home", c.args.home)
	b.scalar("Session key", c.args.key)
	b.scalar("Registered plan file", c.plan.path)
	b.scalar("Registered preparation", c.preparation.path)
	b.scalar("Registered record", c.record.path)
	b.line("Current locator binding is checked only after identity confirmation.")
	b.line("Identity: type confirm " + quote(c.selected.account) + " or cancel:")
	return b
}
func displayCurrent(session reviewcontract.Session, problems []reviewcontract.Diagnostic) builder {
	b := builder{}
	b.line("Current record-backed session.")
	b.line("Instructions are displayed, not executed. Verified is your explicit statement.")
	b.line("Defaults and suggestions are not accepted answers.")
	b.session(session, "", false)
	b.diagnostics("Session problems", problems)
	b.boolean("Plan reference matches current bytes", session.PlanReferenceMatches)
	b.line("Choose one change; nothing here runs a command or check.")
	b.line("answer QUESTION text JSONSTRING")
	b.line("answer QUESTION flag true|false")
	b.line("answer QUESTION list JSONARRAY_OF_STRINGS")
	b.line("answer QUESTION deferred")
	b.line("follow-plan QUESTION")
	b.line("outcome SECTION verified|failed|deferred JSONSTRING_NOTE JSONSTRING_DIFFERENTLY")
	b.line("exit")
	b.line("Action: choose one displayed change or exit:")
	return b
}
func displayProposal(proposal reviewcontract.Proposal, problems []reviewcontract.Diagnostic) builder {
	b := builder{}
	b.line("Selected record change — not saved.")
	b.scalar("Operation", string(proposal.Intent.Operation))
	b.scalar("Question", proposal.Intent.Question)
	b.scalar("Section", proposal.Intent.Section)
	if proposal.Intent.Operation == reviewcontract.OperationOutcome {
		b.scalar("Kind", string(proposal.Intent.Kind))
		var target reviewcontract.Activity
		for _, activity := range proposal.After.Activities {
			if activity.Section.ID == proposal.Intent.Section {
				target = activity
				break
			}
		}
		b.scalar("Current definition hash", target.CurrentHash)
		b.document("Note", proposal.Intent.Note)
		b.document("Differently", proposal.Intent.Differently)
		b.list("Waiting on", target.WaitingOn)
		b.line("New outcome time: sampled after save")
	} else {
		before, after := model.Value{}, model.Value{}
		var scope, name string
		for _, question := range proposal.Before.Questions {
			if question.ID == proposal.Intent.Question {
				before = question.Answer
				scope = question.Scope
				name = question.Name
			}
		}
		for _, question := range proposal.After.Questions {
			if question.ID == proposal.Intent.Question {
				after = question.Answer
			}
		}
		b.value("Before answer", before)
		b.value("After answer", after)
		stored := false
		if scope == "plan" {
			_, stored = proposal.AfterRecord.Answers.Plan[name]
		} else {
			_, stored = proposal.AfterRecord.Answers.Sections[scope][name]
		}
		b.boolean("Stored override after save", stored)
		if proposal.Intent.Operation == reviewcontract.OperationFollowPlan {
			b.line("Following the shared plan may reveal an existing shared answer; it does not promise blank.")
		} else {
			b.line("Explicit submission stores your answer even when equal to a shared answer.")
		}
	}
	b.session(proposal.After, proposal.Intent.Section, proposal.NeedsEvent)
	b.diagnostics("Session problems", problems)
	b.boolean("Plan reference matches current bytes", proposal.After.PlanReferenceMatches)
	b.line("Only the selected record can change; plan, preparation, registration and first-use data are retained.")
	b.line("Changed records use canonical YAML; a no-op retains exact original bytes.")
	b.line("Private staging can remain after interruption; unknown leftovers are preserved, not replayed or automatically deleted.")
	b.line("Review complete. Choose save or cancel:")
	return b
}
func (b *builder) session(s reviewcontract.Session, eventSection string, needsEvent bool) {
	b.count("Activities", len(s.Activities))
	for i, a := range s.Activities {
		b.count("Activity", i+1)
		section := a.Section
		b.scalar("ID", section.ID)
		b.scalar("Title", section.Title)
		b.scalar("Summary", section.Summary)
		b.scalar("Parent", section.Parent)
		b.scalar("Owner", section.Owner)
		b.scalar("Timing", string(section.Timing))
		b.scalar("Verification", section.Verification)
		b.scalar("Current definition hash", a.CurrentHash)
		b.document("Body", section.Body)
		b.list("Requires", section.Requires)
		b.list("After", section.After)
		b.list("Waiting on", a.WaitingOn)
		b.list("Escalation", a.Escalation)
		b.count("Steps", len(section.Steps))
		for j, step := range section.Steps {
			b.count("Step", j+1)
			b.scalar("ID", step.ID)
			b.scalar("Name", step.Name)
			b.scalar("Owner", step.Owner)
			b.scalar("Timing", string(step.Timing))
			b.scalar("Open", step.Open)
			b.scalar("Directory", step.Directory)
			b.document("Notes", step.Notes)
			b.list("Run", step.Run)
			b.list("Escalation", step.Escalation)
			keys := make([]string, 0, len(step.Env))
			for key := range step.Env {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			b.count("Environment", len(keys))
			for _, key := range keys {
				b.add("  - Key: ")
				b.quoted(key)
				b.add(" Value: ")
				b.quoted(step.Env[key])
				b.add("\n")
			}
			b.count("Checks", len(step.Checks))
			for k, check := range step.Checks {
				b.count("Check", k+1)
				b.list("Run", check.Run)
				b.document("Ask", check.Ask)
				b.scalar("Environment question", check.Env)
			}
		}
		if b.failed {
			return
		}
	}
	b.count("Questions", len(s.Questions))
	for i, q := range s.Questions {
		b.count("Question", i+1)
		b.scalar("ID", q.ID)
		b.scalar("Scope", q.Scope)
		b.scalar("Name", q.Name)
		b.scalar("Description", q.Description)
		b.scalar("Example", q.Example)
		b.scalar("Type", string(q.Type))
		b.scalar("Audience", string(q.Audience))
		b.boolean("Required", q.Required)
		b.boolean("Multiple", q.Multiple)
		b.value("Answer", q.Answer)
		b.value("Default suggestion", q.DeclaredDefault)
		b.list("Options", q.Options)
		b.count("Choices", len(q.Choices))
		for _, choice := range q.Choices {
			b.add("  - Value: ")
			b.quoted(choice.Value)
			b.add(" Label: ")
			b.quoted(choice.Label)
			b.add("\n")
		}
		b.count("Suggestions", len(q.Suggestions))
		for j, suggestion := range q.Suggestions {
			b.count("Suggestion", j+1)
			b.scalar("Source", string(suggestion.Source))
			b.value("Value", suggestion.Value)
		}
		if b.failed {
			return
		}
	}
	blockers := make([]string, len(s.Blockers))
	for i, blocker := range s.Blockers {
		blockers[i] = blocker.Question
	}
	b.list("Blockers", blockers)
	b.count("Recorded outcomes", len(s.Outcomes))
	for i, held := range s.Outcomes {
		o := held.Outcome
		b.count("Recorded outcome", i+1)
		b.scalar("Section", o.Section)
		b.scalar("Outcome", string(o.Kind))
		on := string(o.On)
		if needsEvent && o.Section == eventSection {
			on = "sampled after save"
		}
		b.scalar("On", on)
		b.scalar("Recorded definition hash", o.SectionHash)
		b.scalar("Current definition hash", held.CurrentHash)
		b.scalar("Comparison", held.Comparison)
		b.document("Note", o.Note)
		b.document("Differently", o.Differently)
		if b.failed {
			return
		}
	}
	b.boolean("Support available", s.Support.Available)
	if !s.Support.Available {
		b.line("Current support is unavailable; ask the onboarding owner to repair preparation.")
		b.diagnostics("Support problems", s.Support.Problems)
		return
	}
	recipient := s.Support.Recipient
	b.scalar("Name", recipient.Name)
	b.scalar("Company email", recipient.CompanyEmail)
	b.scalar("Planned start date (goal)", string(recipient.PlannedStartDate))
	b.scalar("Team ID", recipient.Team)
	b.scalar("Manager contact ID", recipient.Manager)
	b.scalar("Onboarding owner contact ID", recipient.OnboardingOwner)
	b.scalar("Buddy contact ID", recipient.Buddy)
	b.scalar("Work location", recipient.WorkLocation)
	b.scalar("Time zone", recipient.TimeZone)
	b.count("Arrangements", len(s.Support.Arrangements))
	for i, arrangement := range s.Support.Arrangements {
		b.count("Arrangement", i+1)
		b.scalar("ID", arrangement.ID)
		b.scalar("Section", arrangement.Section)
		b.scalar("Kind", arrangement.Kind)
		b.scalar("Owner contact ID", arrangement.Owner)
		b.scalar("Reference", arrangement.Reference)
		b.scalar("Target date (goal)", string(arrangement.TargetDate))
		b.scalar("Status", arrangement.Status)
		b.list("Helper contact IDs", arrangement.Contacts)
		b.document("Note", arrangement.Note)
		if b.failed {
			return
		}
	}
}
