package validator

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	contract "onboard-kit/contracts/preparation/validator"
	"onboard-kit/internal/model"
)

type validator struct{}

func New() contract.Validator {
	return validator{}
}

func (validator) Validate(context contract.Context, preparation model.Preparation) []model.Problem {
	var problems diagnostics
	if preparation.SchemaVersion != 1 {
		problems.add("version", "schemaVersion", "Use preparation schema version 1.")
		return problems
	}
	if preparation.Kind != "preparation" {
		problems.add("kind", "kind", "Set kind to preparation.")
		return problems
	}
	if !matches(`[A-Za-z0-9][A-Za-z0-9._-]*`, preparation.Account) {
		problems.add("account", "account", "Supply an account beginning with a letter or digit and containing only letters, digits, dots, underscores, or hyphens.")
	} else if preparation.Account != context.Account {
		problems.add("account-mismatch", "account", "Use the expected recipient account.")
	}
	if preparation.PlanID == "" {
		problems.add("plan", "planId", "Supply the organization plan identity.")
	} else if preparation.PlanID != context.PlanID {
		problems.add("plan-mismatch", "planId", "Use the expected organization plan identity.")
	}
	if context.Record != nil {
		if context.Record.Account != context.Account {
			problems.add("record-account", "record.account", "Select a record for the expected recipient account.")
		}
		if context.Record.PlanID != context.PlanID {
			problems.add("record-plan", "record.planId", "Select a record for the expected organization plan.")
		}
	}
	if len(problems) != 0 {
		return problems
	}
	problems.recipient(context, preparation.Recipient)
	seen := make(map[string]bool)
	for index, arrangement := range preparation.Arrangements {
		path := "arrangements[" + strconv.Itoa(index) + "]"
		if !matches(`[a-z][a-z0-9-]*`, arrangement.ID) {
			problems.add("arrangement-id", path+".id", "Supply an arrangement id beginning with a lowercase letter and containing only lowercase letters, digits, or hyphens.")
		} else {
			if seen[arrangement.ID] {
				problems.add("duplicate-id", path+".id", "Give each arrangement a unique id.")
			}
			seen[arrangement.ID] = true
		}
		problems.arrangement(context, arrangement, path)
	}
	return problems
}

type diagnostics []model.Problem

func (problems *diagnostics) add(code, path, message string) {
	*problems = append(*problems, model.Problem{
		Code:     "preparation." + code,
		Severity: model.SeverityError,
		Message:  message,
		Detail:   map[string]string{"path": path},
	})
}

func (problems *diagnostics) date(date model.Date, path string) {
	if date == "" {
		return
	}
	parsed, err := time.Parse(time.DateOnly, string(date))
	if err != nil || parsed.Format(time.DateOnly) != string(date) {
		problems.add("date", path, "Supply a real calendar date in YYYY-MM-DD format.")
	}
}

func (problems *diagnostics) contact(contacts []string, value, path string) {
	if !slices.Contains(contacts, value) || strings.TrimSpace(value) == "" {
		problems.add("unknown-contact", path, "Choose a declared contact id.")
	}
}

func (problems *diagnostics) recipient(context contract.Context, recipient model.PreparationRecipient) {
	problems.date(recipient.PlannedStartDate, "recipient.plannedStartDate")
	if recipient.Team != "" && !slices.Contains(context.Teams, recipient.Team) {
		problems.add("unknown-team", "recipient.team", "Choose a declared team id.")
	}
	for _, field := range []struct{ name, value string }{
		{"manager", recipient.Manager},
		{"onboardingOwner", recipient.OnboardingOwner},
		{"buddy", recipient.Buddy},
	} {
		if field.value != "" {
			problems.contact(context.Contacts, field.value, "recipient."+field.name)
		}
	}
	switch recipient.WorkLocation {
	case "", "remote", "on-site", "hybrid":
	default:
		problems.add("location", "recipient.workLocation", "Choose remote, on-site, or hybrid, or leave work location empty.")
	}
}

func (problems *diagnostics) arrangement(context contract.Context, arrangement model.Arrangement, path string) {
	if !slices.Contains(context.Sections, arrangement.Section) {
		problems.add("unknown-section", path+".section", "Choose a declared section id.")
	}
	known, needsContacts := arrangementKind(arrangement.Kind)
	if !known {
		problems.add("arrangement-kind", path+".kind", "Choose a supported preparation arrangement kind.")
	}
	switch arrangement.Status {
	case "pending", "requested", "arranged", "deferred":
	default:
		problems.add("arrangement-status", path+".status", "Set status to pending, requested, arranged, or deferred.")
	}
	if arrangement.Owner != "" {
		problems.contact(context.Contacts, arrangement.Owner, path+".owner")
	} else if arrangement.Status == "arranged" {
		problems.add("incomplete-arrangement", path+".owner", "Assign an owner before marking the arrangement arranged.")
	}
	if arrangement.Status == "arranged" && needsContacts && len(arrangement.Contacts) == 0 {
		problems.add("incomplete-arrangement", path+".contacts", "Assign a helper before marking this arrangement arranged.")
	}
	for index, contact := range arrangement.Contacts {
		problems.contact(context.Contacts, contact, path+".contacts["+strconv.Itoa(index)+"]")
	}
	if arrangement.Status == "arranged" && known && arrangement.Kind != "support" && strings.TrimSpace(arrangement.Reference) == "" {
		problems.add("incomplete-arrangement", path+".reference", "Supply a work or resource reference before marking this arrangement arranged.")
	}
	problems.date(arrangement.TargetDate, path+".targetDate")
}

func arrangementKind(kind string) (known, needsContacts bool) {
	switch kind {
	case "support", "equipment", "office", "remote", "access", "meeting", "reading":
		return true, false
	case "pairing", "ticket", "pr-review", "pr-post", "pr-merge", "deploy-shadow", "deploy-lead":
		return true, true
	default:
		return false, false
	}
}

func matches(pattern, value string) bool {
	matched, _ := regexp.MatchString("^(?:"+pattern+")$", value)
	return matched
}
