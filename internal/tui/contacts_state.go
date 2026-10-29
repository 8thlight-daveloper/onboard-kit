package tui

import (
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

type directoryKind int

const (
	kindContact directoryKind = iota
	kindRole
	kindTeam
)

const unknownContactText = "someone who is not a contact"

var directoryKinds = []directoryKind{kindContact, kindRole, kindTeam}

func (kind directoryKind) singular() string {
	switch kind {
	case kindRole:
		return "role"
	case kindTeam:
		return "team"
	}
	return "contact"
}

func (kind directoryKind) plural() string { return kind.singular() + "s" }

func (kind directoryKind) title() string {
	return strings.ToUpper(kind.plural()[:1]) + kind.plural()[1:]
}

func kindOfQuestion(question model.Question) (directoryKind, bool) {
	switch question.Type {
	case model.VariableTypeContact:
		return kindContact, true
	case model.VariableTypeRole:
		return kindRole, true
	case model.VariableTypeTeam:
		return kindTeam, true
	}
	return kindContact, false
}

type directoryEntry struct {
	id   string
	name string
}

func entriesOf(directory model.Directory, kind directoryKind) []directoryEntry {
	var entries []directoryEntry
	switch kind {
	case kindContact:
		for _, contact := range directory.Contacts {
			entries = append(entries, directoryEntry{contact.ID, contact.Name})
		}
	case kindRole:
		for _, role := range directory.Roles {
			entries = append(entries, directoryEntry{role.ID, role.Name})
		}
	case kindTeam:
		for _, team := range directory.Teams {
			entries = append(entries, directoryEntry{team.ID, team.Name})
		}
	}
	return entries
}

func entryName(directory model.Directory, key entryKey) string {
	for _, entry := range entriesOf(directory, key.kind) {
		if entry.id == key.id {
			return entry.name
		}
	}
	return key.id
}

func idsOf(entries []directoryEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.id)
	}
	return ids
}

func contactChoices(directory model.Directory) []model.Choice {
	choices := make([]model.Choice, 0, len(directory.Contacts))
	for _, contact := range directory.Contacts {
		choices = append(choices, model.Choice{Value: contact.ID, Label: contact.Name})
	}
	return choices
}

func holderNames(directory model.Directory, ids []string) []string {
	choices := contactChoices(directory)
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if label, ok := labelOf(choices, id); ok {
			names = append(names, label)
			continue
		}
		names = append(names, unknownContactText)
	}
	return names
}

func roleNameOf(directory model.Directory, id string) string {
	for _, role := range directory.Roles {
		if role.ID == id {
			return role.Name
		}
	}
	return "a role that does not exist"
}

func teamOverrideRoles(directory model.Directory, team model.Team) []string {
	var roles []string
	for _, role := range directory.Roles {
		if _, ok := team.Holders[role.ID]; ok {
			roles = append(roles, role.ID)
		}
	}
	var unknown []string
	for role := range team.Holders {
		if !slices.Contains(roles, role) {
			unknown = append(unknown, role)
		}
	}
	slices.Sort(unknown)
	return append(roles, unknown...)
}

type contactsMode int

const (
	contactsBrowsing contactsMode = iota
	contactsEditing
	contactsRemoving
)

type entryKey struct {
	kind directoryKind
	id   string
}

type contactsState struct {
	kind    directoryKind
	cursors [3]int
	mode    contactsMode
	form    formState
	removal removalState
	flags   map[entryKey][]model.Problem
}

func (state contactsState) cursor() int { return state.cursors[state.kind] }

func (state contactsState) moved(delta int, count int) contactsState {
	state.cursors[state.kind] = max(min(state.cursors[state.kind]+delta, count-1), 0)
	return state
}

func (state contactsState) cycled(delta int) contactsState {
	count := len(directoryKinds)
	state.kind = directoryKind((int(state.kind) + delta + count) % count)
	return state
}

func (state contactsState) clamped(directory model.Directory) contactsState {
	for _, kind := range directoryKinds {
		state.cursors[kind] = max(min(state.cursors[kind], len(entriesOf(directory, kind))-1), 0)
	}
	return state
}

func (state contactsState) pointedAt(directory model.Directory, kind directoryKind, id string) contactsState {
	state.kind = kind
	if index := slices.Index(idsOf(entriesOf(directory, kind)), id); index >= 0 {
		state.cursors[kind] = index
	}
	return state
}

func (state contactsState) typing() bool { return state.mode == contactsEditing }

func (app App) selectedEntry() (directoryEntry, bool) {
	entries := entriesOf(app.plan.Directory, app.contacts.kind)
	if app.contacts.cursor() >= len(entries) {
		return directoryEntry{}, false
	}
	return entries[app.contacts.cursor()], true
}

func (app App) canEditDirectory() bool { return app.audience != model.AudienceOnboardee }

func directoryFlags(problems []model.Problem) map[entryKey][]model.Problem {
	flags := map[entryKey][]model.Problem{}
	for _, problem := range problems {
		if !strings.HasPrefix(problem.Code, "directory.") || problem.Question != "" {
			continue
		}
		if key, ok := flagKey(problem); ok {
			flags[key] = append(flags[key], problem)
		}
	}
	return flags
}

func flagKey(problem model.Problem) (entryKey, bool) {
	detail := problem.Detail
	switch problem.Code {
	case "directory.contact-unreachable", "directory.email-malformed":
		return entryKey{kindContact, detail["contact"]}, true
	case "directory.role-vacant":
		return entryKey{kindRole, detail["role"]}, true
	case "directory.team-role-unknown":
		return entryKey{kindTeam, detail["team"]}, true
	case "directory.holder-not-a-contact":
		if team, ok := detail["team"]; ok {
			return entryKey{kindTeam, team}, true
		}
		return entryKey{kindRole, detail["role"]}, true
	case "directory.duplicate-id":
		return entryKey{kindNamed(detail["kind"]), detail["id"]}, true
	}
	return entryKey{}, false
}

func kindNamed(name string) directoryKind {
	for _, kind := range directoryKinds {
		if kind.singular() == name {
			return kind
		}
	}
	return kindContact
}

func isTeamContactProblem(problem model.Problem) bool {
	_, ok := problem.Detail["team-contact"]
	return ok
}

func flagWords(problem model.Problem) string {
	switch problem.Code {
	case "directory.role-vacant":
		return "vacant"
	case "directory.contact-unreachable":
		return "no way to reach"
	case "directory.email-malformed":
		return "email looks wrong"
	case "directory.holder-not-a-contact":
		if isTeamContactProblem(problem) {
			return "team contact is not a contact"
		}
		return "holder is not a contact"
	case "directory.team-role-unknown":
		return "role does not exist"
	case "directory.duplicate-id":
		return "id is shared"
	}
	return "needs a look"
}

func flagSentence(problem model.Problem) string {
	switch problem.Code {
	case "directory.role-vacant":
		return "Nobody holds this role, so it shows as a gap in the document."
	case "directory.contact-unreachable":
		return "No email, Slack handle, or Slack id, so nobody can reach this contact."
	case "directory.email-malformed":
		return "The email needs an @ and no spaces."
	case "directory.holder-not-a-contact":
		if isTeamContactProblem(problem) {
			return "The contact for this team is not one of the contacts. Pick another, or choose no team contact."
		}
		return "A holder is not one of the contacts."
	case "directory.team-role-unknown":
		return "This team overrides a role that does not exist."
	case "directory.duplicate-id":
		return "Another entry of this kind has the same id."
	}
	return problem.Message
}

func (app App) flagsOf(kind directoryKind, id string) []model.Problem {
	return app.contacts.flags[entryKey{kind, id}]
}

func (app App) refreshedContacts() App {
	app.contacts.flags = directoryFlags(app.backend.Validate(app.plan))
	app.contacts = app.contacts.clamped(app.plan.Directory)
	return app
}
