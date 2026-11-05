package tui

import (
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	"onboard-kit/internal/model"
)

const noTeamContactText = "No team contact"

const (
	formLabelWidth = 10
	formMarkWidth  = 2
)

type overrideRow struct {
	roleID   string
	roleName string
	special  int
	holders  []string
}

type formState struct {
	kind        directoryKind
	editing     string
	taken       []string
	contacts    []model.Choice
	labels      []string
	hints       []string
	fields      []textinput.Model
	focus       int
	holders     picker
	rows        []overrideRow
	choosing    bool
	chooser     picker
	teamContact string
	problems    []model.Problem
}

func newForm(directory model.Directory, kind directoryKind, id string, width int) formState {
	form := formState{kind: kind, editing: id, taken: idsOf(entriesOf(directory, kind)), contacts: contactChoices(directory)}
	switch kind {
	case kindContact:
		form = form.withContactFields(directory, id)
	case kindRole:
		form = form.withRoleFields(directory, id)
	case kindTeam:
		form = form.withTeamFields(directory, id)
	}
	return form.focused(0).withWidth(width)
}

func newTextField(value string) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetValue(value)
	input.CursorEnd()
	return input
}

func (form formState) withContactFields(directory model.Directory, id string) formState {
	var existing model.Contact
	for _, contact := range directory.Contacts {
		if contact.ID == id {
			existing = contact
		}
	}
	form.labels = []string{"Name", "Email", "Slack", "Slack id"}
	form.hints = []string{"e.g. Sam Okafor", "optional", "optional, such as @sam or #it-help", "optional, used for links"}
	form.fields = []textinput.Model{
		newTextField(existing.Name),
		newTextField(existing.Email),
		newTextField(existing.Slack),
		newTextField(existing.SlackID),
	}
	return form
}

func (form formState) withRoleFields(directory model.Directory, id string) formState {
	var existing model.Role
	for _, role := range directory.Roles {
		if role.ID == id {
			existing = role
		}
	}
	form.labels = []string{"Name"}
	form.hints = []string{"e.g. Engineering manager"}
	form.fields = []textinput.Model{newTextField(existing.Name)}
	form.holders = newPicker(nil, contactChoices(directory), existing.Holders, specialNone)
	return form
}

func (form formState) withTeamFields(directory model.Directory, id string) formState {
	var existing model.Team
	for _, team := range directory.Teams {
		if team.ID == id {
			existing = team
		}
	}
	if _, ok := labelOf(contactChoices(directory), existing.Contact); ok {
		form.teamContact = existing.Contact
	}
	form.labels = []string{"Name"}
	form.hints = []string{"e.g. Platform team"}
	form.fields = []textinput.Model{newTextField(existing.Name)}
	for _, role := range directory.Roles {
		row := overrideRow{roleID: role.ID, roleName: role.Name, special: specialInherit}
		if holders, ok := existing.Holders[role.ID]; ok {
			row.special, row.holders = specialVacant, nil
			if len(holders) > 0 {
				row.special, row.holders = specialNone, knownHolders(directory, holders)
			}
		}
		form.rows = append(form.rows, row)
	}
	return form
}

func knownHolders(directory model.Directory, ids []string) []string {
	var known []string
	for _, id := range ids {
		if _, ok := labelOf(contactChoices(directory), id); ok {
			known = append(known, id)
		}
	}
	return known
}

func (form formState) withWidth(width int) formState {
	fields := slices.Clone(form.fields)
	for index := range fields {
		fields[index].Width = max(width-formMarkWidth-formLabelWidth-1, inputMinimum)
		fields[index].Placeholder = truncateCells(form.hints[index], fields[index].Width)
		fields[index].SetCursor(fields[index].Position())
	}
	form.fields = fields
	return form
}

func (form formState) focused(index int) formState {
	form.focus = max(min(index, form.lastFocus()), 0)
	fields := slices.Clone(form.fields)
	for position := range fields {
		fields[position].Blur()
	}
	if form.focus < len(fields) {
		fields[form.focus].Focus()
	}
	form.fields = fields
	return form
}

func (form formState) withoutFocus() formState {
	fields := slices.Clone(form.fields)
	for position := range fields {
		fields[position].Blur()
	}
	form.fields = fields
	return form
}

func (form formState) lastFocus() int {
	switch form.kind {
	case kindRole:
		return len(form.fields)
	case kindTeam:
		return len(form.fields) + len(form.rows)
	}
	return len(form.fields) - 1
}

func (form formState) onText() bool { return form.focus < len(form.fields) }

func (form formState) onHolders() bool {
	return form.kind == kindRole && form.focus == len(form.fields)
}

func (form formState) onRows() bool {
	return form.kind == kindTeam && form.focus >= len(form.fields)
}

func (form formState) onContactRow() bool {
	return form.kind == kindTeam && form.focus == len(form.fields)
}

func (form formState) rowIndex() int { return form.focus - len(form.fields) - 1 }

func (form formState) name() string { return strings.TrimSpace(form.fields[0].Value()) }

func (form formState) id() string {
	if form.editing != "" {
		return form.editing
	}
	return freshID(form.name(), form.taken)
}

func (form formState) openedPicker() formState {
	if !form.onRows() {
		return form
	}
	if form.onContactRow() {
		form.chooser = newPicker([]string{noTeamContactText}, form.contacts, contactPicks(form.teamContact), contactSpecial(form.teamContact)).onlyOne()
		form.choosing = true
		return form
	}
	row := form.rows[form.rowIndex()]
	form.chooser = newPicker([]string{"Use the role's own holders", "Vacant for this team"}, form.contacts, row.holders, row.special)
	form.choosing = true
	return form
}

func (form formState) closedPicker(keep bool) formState {
	if keep && form.onContactRow() {
		form.teamContact = ""
		if len(form.chooser.picked) > 0 {
			form.teamContact = form.chooser.picked[0]
		}
	} else if keep {
		rows := slices.Clone(form.rows)
		rows[form.rowIndex()].special = form.chooser.special
		rows[form.rowIndex()].holders = form.chooser.picked
		form.rows = rows
	}
	form.choosing = false
	return form
}

func (form formState) contact() model.Contact {
	return model.Contact{
		ID:      form.id(),
		Name:    form.name(),
		Email:   strings.TrimSpace(form.fields[1].Value()),
		Slack:   strings.TrimSpace(form.fields[2].Value()),
		SlackID: strings.TrimSpace(form.fields[3].Value()),
	}
}

func (form formState) role() model.Role {
	return model.Role{ID: form.id(), Name: form.name(), Holders: slices.Clone(form.holders.picked)}
}

func (form formState) team() model.Team {
	team := model.Team{ID: form.id(), Name: form.name(), Contact: form.teamContact}
	for _, row := range form.rows {
		switch row.special {
		case specialVacant:
			team.Holders = withOverride(team.Holders, row.roleID, []string{})
		case specialNone:
			team.Holders = withOverride(team.Holders, row.roleID, slices.Clone(row.holders))
		}
	}
	return team
}

func contactPicks(contact string) []string {
	if contact == "" {
		return nil
	}
	return []string{contact}
}

func contactSpecial(contact string) int {
	if contact == "" {
		return specialInherit
	}
	return specialNone
}

func withOverride(overrides map[string][]string, role string, holders []string) map[string][]string {
	if overrides == nil {
		overrides = map[string][]string{}
	}
	overrides[role] = holders
	return overrides
}

func slugOf(name string) string {
	var slug strings.Builder
	pendingHyphen := false
	for _, character := range strings.ToLower(name) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			if pendingHyphen && slug.Len() > 0 {
				slug.WriteByte('-')
			}
			pendingHyphen = false
			slug.WriteRune(character)
			continue
		}
		pendingHyphen = true
	}
	return slug.String()
}

func freshID(name string, taken []string) string {
	base := slugOf(name)
	if base == "" || !slices.Contains(taken, base) {
		return base
	}
	for number := 2; ; number++ {
		candidate := base + "-" + strconv.Itoa(number)
		if !slices.Contains(taken, candidate) {
			return candidate
		}
	}
}

func plainProblem(problem model.Problem) string {
	switch problem.Code {
	case "directory.id-malformed":
		return "The name needs a letter or a digit, so an id can be made from it."
	case "directory.name-empty":
		return "A name is needed."
	case "directory.holder-not-a-contact":
		if isTeamContactProblem(problem) {
			return "The team contact is not one of the contacts."
		}
		return "A holder is not one of the contacts."
	case "directory.team-role-unknown":
		return "A team overrides a role that is not there."
	case "directory.entry-not-found":
		return "That entry is not in the directory any more."
	}
	return problem.Message
}

func plainProblems(problems []model.Problem, name string) []model.Problem {
	var plain []model.Problem
	for _, problem := range problems {
		if problem.Code == "directory.id-malformed" && name == "" {
			continue
		}
		problem.Message = plainProblem(problem)
		plain = append(plain, problem)
	}
	return plain
}
