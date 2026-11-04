package tui

import (
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

type valueKind int

const (
	valueText valueKind = iota
	valueDate
	valueContact
	valueTeam
	valueSection
	valueArrangementKind
	valueStatus
	valueLocation
	valueHelpers
)

const (
	notSetText         = "not set"
	notChosenText      = "not chosen"
	unknownTeamText    = "a team that does not exist"
	unknownSectionText = "a section that does not exist"
)

type fieldSpec[T any] struct {
	label string
	hint  string
	kind  valueKind
	read  func(T) []string
	write func(T, []string) T
}

func single(text string) []string {
	if text == "" {
		return nil
	}
	return []string{text}
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func textSingle(text string) []string { return []string{text} }

func recipientSpecs() []fieldSpec[model.PreparationRecipient] {
	type recipient = model.PreparationRecipient
	return []fieldSpec[recipient]{
		{"Name", "e.g. Alex Doe", valueText, func(held recipient) []string { return textSingle(held.Name) }, func(held recipient, values []string) recipient { held.Name = first(values); return held }},
		{"Company email", "optional", valueText, func(held recipient) []string { return textSingle(held.CompanyEmail) }, func(held recipient, values []string) recipient { held.CompanyEmail = first(values); return held }},
		{"Planned start", "optional, YYYY-MM-DD", valueDate, func(held recipient) []string { return textSingle(string(held.PlannedStartDate)) }, func(held recipient, values []string) recipient {
			held.PlannedStartDate = model.Date(first(values))
			return held
		}},
		{"Team", "", valueTeam, func(held recipient) []string { return single(held.Team) }, func(held recipient, values []string) recipient { held.Team = first(values); return held }},
		{"Manager", "", valueContact, func(held recipient) []string { return single(held.Manager) }, func(held recipient, values []string) recipient { held.Manager = first(values); return held }},
		{"Onboarding owner", "", valueContact, func(held recipient) []string { return single(held.OnboardingOwner) }, func(held recipient, values []string) recipient { held.OnboardingOwner = first(values); return held }},
		{"Buddy", "", valueContact, func(held recipient) []string { return single(held.Buddy) }, func(held recipient, values []string) recipient { held.Buddy = first(values); return held }},
		{"Work location", "", valueLocation, func(held recipient) []string { return single(held.WorkLocation) }, func(held recipient, values []string) recipient { held.WorkLocation = first(values); return held }},
		{"Time zone", "optional, such as Europe/Berlin", valueText, func(held recipient) []string { return textSingle(held.TimeZone) }, func(held recipient, values []string) recipient { held.TimeZone = first(values); return held }},
	}
}

func arrangementSpecs() []fieldSpec[model.Arrangement] {
	type arrangement = model.Arrangement
	return []fieldSpec[arrangement]{
		{"Section", "", valueSection, func(held arrangement) []string { return single(held.Section) }, func(held arrangement, values []string) arrangement { held.Section = first(values); return held }},
		{"Kind", "", valueArrangementKind, func(held arrangement) []string { return single(held.Kind) }, func(held arrangement, values []string) arrangement { held.Kind = first(values); return held }},
		{"Owner", "", valueContact, func(held arrangement) []string { return single(held.Owner) }, func(held arrangement, values []string) arrangement { held.Owner = first(values); return held }},
		{"Helpers", "", valueHelpers, func(held arrangement) []string { return slices.Clone(held.Contacts) }, func(held arrangement, values []string) arrangement { held.Contacts = slices.Clone(values); return held }},
		{"Reference", "the ticket, pull request, or resource", valueText, func(held arrangement) []string { return textSingle(held.Reference) }, func(held arrangement, values []string) arrangement { held.Reference = first(values); return held }},
		{"Target date", "optional, YYYY-MM-DD", valueDate, func(held arrangement) []string { return textSingle(string(held.TargetDate)) }, func(held arrangement, values []string) arrangement {
			held.TargetDate = model.Date(first(values))
			return held
		}},
		{"Status", "", valueStatus, func(held arrangement) []string { return single(held.Status) }, func(held arrangement, values []string) arrangement { held.Status = first(values); return held }},
		{"Note", "optional, no secrets or private details", valueText, func(held arrangement) []string { return textSingle(held.Note) }, func(held arrangement, values []string) arrangement { held.Note = first(values); return held }},
	}
}

func (kind valueKind) typed() bool { return kind == valueText || kind == valueDate }

func (kind valueKind) picksMany() bool { return kind == valueHelpers }

func namedChoices(names ...string) []model.Choice {
	choices := make([]model.Choice, 0, len(names))
	for _, name := range names {
		choices = append(choices, model.Choice{Value: name, Label: name})
	}
	return choices
}

func choicesOfKind(kind valueKind, plan model.Plan) []model.Choice {
	switch kind {
	case valueContact, valueHelpers:
		return contactChoices(plan.Directory)
	case valueTeam:
		choices := make([]model.Choice, 0, len(plan.Directory.Teams))
		for _, team := range plan.Directory.Teams {
			choices = append(choices, model.Choice{Value: team.ID, Label: team.Name})
		}
		return choices
	case valueSection:
		choices := make([]model.Choice, 0, len(plan.Sections))
		for _, section := range plan.Sections {
			label := section.Title
			if !isOn(section) {
				label += " (switched off)"
			}
			choices = append(choices, model.Choice{Value: section.ID, Label: label})
		}
		return choices
	case valueArrangementKind:
		return namedChoices("support", "equipment", "office", "remote", "access", "meeting", "reading", "pairing", "ticket", "pr-review", "pr-post", "pr-merge", "deploy-shadow", "deploy-lead")
	case valueStatus:
		return namedChoices("pending", "requested", "arranged", "deferred")
	case valueLocation:
		return namedChoices("remote", "on-site", "hybrid")
	}
	return nil
}

func leadingOfKind(kind valueKind) []string {
	switch kind {
	case valueContact:
		return []string{"Nobody"}
	case valueTeam:
		return []string{"No team"}
	case valueLocation:
		return []string{"Not set"}
	}
	return nil
}

func emptyWords(kind valueKind) string {
	switch kind {
	case valueSection, valueArrangementKind, valueStatus:
		return notChosenText
	case valueHelpers:
		return "nobody"
	}
	return notSetText
}

func displayValues(kind valueKind, values []string, plan model.Plan) string {
	if len(values) == 0 || kind.typed() && values[0] == "" {
		return emptyWords(kind)
	}
	switch kind {
	case valueContact, valueHelpers:
		return strings.Join(holderNames(plan.Directory, values), ", ")
	case valueTeam, valueSection:
		if label, ok := labelOf(choicesOfKind(kind, plan), values[0]); ok {
			return label
		}
		if kind == valueTeam {
			return unknownTeamText
		}
		return unknownSectionText
	}
	return visibleText(singleLine(values[0]))
}

func quotedValues(kind valueKind, values []string, plan model.Plan) string {
	shown := displayValues(kind, values, plan)
	if kind == valueText && len(values) > 0 && values[0] != "" {
		return "\"" + shown + "\""
	}
	return shown
}
