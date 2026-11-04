package tui

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	"onboard-kit/internal/model"
)

const preparationLabelWidth = 18

type preparationField struct {
	label  string
	hint   string
	kind   valueKind
	held   []string
	input  textinput.Model
	shown  string
	picked []string
	chosen bool
}

type preparationForm struct {
	arrangement bool
	editing     string
	taken       []string
	plan        model.Plan
	fields      []preparationField
	focus       int
	choosing    bool
	chooser     picker
	problems    []string
}

func newPreparationField[T any](spec fieldSpec[T], held T) preparationField {
	field := preparationField{label: spec.label, hint: spec.hint, kind: spec.kind, held: spec.read(held)}
	if spec.kind.typed() {
		field.shown = visibleText(singleLine(first(field.held)))
		field.input = newTextField(field.shown)
	}
	return field
}

func idField(id string) preparationField {
	spec := fieldSpec[string]{label: "Id", hint: "e.g. first-pairing, which never changes", kind: valueText, read: textSingle}
	return newPreparationField(spec, id)
}

func recipientForm(person openPerson, width int) preparationForm {
	form := preparationForm{plan: person.loaded.Plan}
	for _, spec := range recipientSpecs() {
		form.fields = append(form.fields, newPreparationField(spec, person.draft.Recipient))
	}
	return form.focused(0).withWidth(width)
}

func arrangementForm(person openPerson, index int, width int) preparationForm {
	form := preparationForm{arrangement: true, plan: person.loaded.Plan}
	var held model.Arrangement
	if index >= 0 && index < len(person.arrangements()) {
		held = person.arrangements()[index]
		form.editing = held.ID
	} else {
		form.fields = append(form.fields, idField(""))
		for _, arrangement := range person.arrangements() {
			form.taken = append(form.taken, arrangement.ID)
		}
	}
	for _, spec := range arrangementSpecs() {
		form.fields = append(form.fields, newPreparationField(spec, held))
	}
	return form.focused(0).withWidth(width)
}

func (field preparationField) values() []string {
	switch {
	case field.kind.typed() && field.input.Value() == field.shown:
		return field.held
	case field.kind.typed():
		return []string{field.input.Value()}
	case field.chosen:
		return slices.Clone(field.picked)
	}
	return field.held
}

func (form preparationForm) withWidth(width int) preparationForm {
	fields := slices.Clone(form.fields)
	for index := range fields {
		if !fields[index].kind.typed() {
			continue
		}
		fields[index].input.Width = max(width-formMarkWidth-preparationLabelWidth-1, inputMinimum)
		fields[index].input.Placeholder = truncateCells(fields[index].hint, fields[index].input.Width)
		fields[index].input.SetCursor(fields[index].input.Position())
	}
	form.fields = fields
	return form
}

func (form preparationForm) focused(index int) preparationForm {
	form.focus = max(min(index, len(form.fields)-1), 0)
	fields := slices.Clone(form.fields)
	for position := range fields {
		if fields[position].kind.typed() {
			fields[position].input.Blur()
		}
	}
	if fields[form.focus].kind.typed() {
		fields[form.focus].input.Focus()
	}
	form.fields = fields
	return form
}

func (form preparationForm) current() preparationField { return form.fields[form.focus] }

func (form preparationForm) onText() bool { return form.current().kind.typed() }

func (form preparationForm) onLast() bool { return form.focus == len(form.fields)-1 }

func (form preparationForm) openedPicker() preparationForm {
	field := form.current()
	if field.kind.typed() {
		return form
	}
	values := field.values()
	special := specialNone
	leading := leadingOfKind(field.kind)
	if len(leading) > 0 && len(values) == 0 {
		special = specialInherit
	}
	form.chooser = newPicker(leading, choicesOfKind(field.kind, form.plan), values, special)
	if !field.kind.picksMany() {
		form.chooser = form.chooser.onlyOne()
	}
	form.choosing = true
	return form
}

func (form preparationForm) closedPicker(keep bool) preparationForm {
	if keep {
		fields := slices.Clone(form.fields)
		field := fields[form.focus]
		field.chosen = true
		field.picked = slices.Clone(form.chooser.picked)
		if form.chooser.special >= 0 && len(form.chooser.leading) > 0 {
			field.picked = nil
		}
		fields[form.focus] = field
		form.fields = fields
	}
	form.choosing = false
	return form
}

func formValue[T any](fields []preparationField, specs []fieldSpec[T], start T) T {
	value := start
	for index, spec := range specs {
		value = spec.write(value, fields[index].values())
	}
	return value
}

func (form preparationForm) specFields() []preparationField {
	if form.arrangement && form.editing == "" {
		return form.fields[1:]
	}
	return form.fields
}

func (form preparationForm) recipient() model.PreparationRecipient {
	return formValue(form.fields, recipientSpecs(), model.PreparationRecipient{})
}

func (form preparationForm) arrangementValue() model.Arrangement {
	id := form.editing
	if id == "" {
		id = first(form.fields[0].values())
	}
	return formValue(form.specFields(), arrangementSpecs(), model.Arrangement{ID: id})
}

func (form preparationForm) change() preparationChange {
	if form.arrangement {
		return preparationChange{kind: changeArrangement, arrangement: form.arrangementValue()}
	}
	return preparationChange{kind: changeRecipient, recipient: form.recipient()}
}

func (form preparationForm) idProblem() (string, bool) {
	if !form.arrangement || form.editing != "" {
		return "", false
	}
	id := form.arrangementValue().ID
	if strings.TrimSpace(id) == "" {
		return "Id: An id is needed, such as first-pairing.", true
	}
	if slices.Contains(form.taken, id) {
		return "Id: " + id + " is taken. To change that arrangement, choose it on the list.", true
	}
	return "", false
}

func (form preparationForm) labelled(problems []model.Problem) []string {
	var shown []string
	for _, problem := range problems {
		shown = append(shown, form.labelOfPath(problem.Detail["path"])+problem.Message)
	}
	return shown
}

func (form preparationForm) labelOfPath(path string) string {
	name := path[strings.LastIndex(path, ".")+1:]
	if bracket := strings.Index(name, "["); bracket >= 0 {
		name = name[:bracket]
	}
	labels := map[string]string{
		"name": "Name", "companyEmail": "Company email", "plannedStartDate": "Planned start", "team": "Team",
		"manager": "Manager", "onboardingOwner": "Onboarding owner", "buddy": "Buddy", "workLocation": "Work location",
		"timeZone": "Time zone", "id": "Id", "section": "Section", "kind": "Kind", "owner": "Owner", "contacts": "Helpers",
		"reference": "Reference", "targetDate": "Target date", "status": "Status", "note": "Note",
	}
	if label, ok := labels[name]; ok {
		return label + ": "
	}
	return ""
}
