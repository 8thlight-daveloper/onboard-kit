package guide

import (
	"strconv"

	"onboard-kit/internal/model"
)

var recipientLabels = [...]string{"Name", "Company email", "Planned start date (goal)", "Team ID", "Manager contact ID", "Onboarding owner contact ID", "Buddy contact ID", "Work location", "Time zone"}

func recipientValues(value model.PreparationRecipient) [9]string {
	return [9]string{value.Name, value.CompanyEmail, string(value.PlannedStartDate), value.Team, value.Manager, value.OnboardingOwner, value.Buddy, value.WorkLocation, value.TimeZone}
}

func recipientFrom(values [9]string) model.PreparationRecipient {
	return model.PreparationRecipient{Name: values[0], CompanyEmail: values[1], PlannedStartDate: model.Date(values[2]), Team: values[3], Manager: values[4], OnboardingOwner: values[5], Buddy: values[6], WorkLocation: values[7], TimeZone: values[8]}
}

func (s *session) recipientForm() {
	form := recipientValues(s.draft.Recipient)
	for s.active() {
		for index, label := range recipientLabels {
			s.line(strconv.Itoa(index+1) + ". " + label + ": " + quote(form[index]))
		}
		action, ok := s.ask("Recipient form — not saved. Choose 1-9, apply, or back:")
		if !ok || action == "back" {
			return
		}
		if action == "apply" {
			updated, problems := s.editor.SetRecipient(s.context, s.draft, recipientFrom(form))
			if len(problems) == 0 {
				s.draft = updated
				return
			}
			s.line("Recipient not applied.")
			s.problems(problems)
			continue
		}
		index, valid := ordinal(action, 9)
		if !valid {
			s.invalid()
			continue
		}
		var choices func()
		if index == 3 {
			choices = s.teams
		}
		if index >= 4 && index <= 6 {
			choices = s.contacts
		}
		form[index] = s.scalar(recipientLabels[index], form[index], choices)
	}
}

var arrangementLabels = [...]string{"Section ID", "Kind", "Accountable owner contact ID", "Assigned contacts", "Reference", "Target date (goal)", "Preparation status", "Note"}

func arrangementValues(value model.Arrangement) [8]string {
	return [8]string{value.Section, value.Kind, value.Owner, "", value.Reference, string(value.TargetDate), value.Status, value.Note}
}

func arrangementFrom(id string, values [8]string, contacts []string) model.Arrangement {
	return model.Arrangement{ID: id, Section: values[0], Kind: values[1], Owner: values[2], Contacts: contacts, Reference: values[4], TargetDate: model.Date(values[5]), Status: values[6], Note: values[7]}
}

func (s *session) arrangements() {
	for s.active() {
		s.line("Arrangements — draft not saved.")
		if len(s.draft.Arrangements) == 0 {
			s.line("No arrangements.")
		}
		for index, item := range s.draft.Arrangements {
			s.line(strconv.Itoa(index+1) + ". ID: " + quote(item.ID) + " Section: " + quote(item.Section) + " Kind: " + quote(item.Kind) + " Status: " + quote(item.Status))
		}
		action, ok := s.ask("Arrangements — not saved. Choose new, edit, remove, or back:")
		if !ok || action == "back" {
			return
		}
		switch action {
		case "new":
			s.arrangementForm(model.Arrangement{}, true)
		case "edit", "remove":
			index, selected := s.selection("Choose arrangement number or back:", len(s.draft.Arrangements))
			if !selected {
				continue
			}
			if action == "edit" {
				s.arrangementForm(s.draft.Arrangements[index], false)
			} else {
				s.removal(index)
			}
		default:
			s.invalid()
		}
	}
}

func (s *session) arrangementForm(held model.Arrangement, create bool) {
	id := held.ID
	fields := arrangementValues(held)
	contacts := copyContacts(held.Contacts)
	for s.active() {
		mode := " [fixed]"
		prompt := "Arrangement form — not saved. Choose 1-8, apply, or back:"
		if create {
			mode = " [new]"
			prompt = "Arrangement form — not saved. Choose id, 1-8, apply, or back:"
		}
		s.line("Permanent ID: " + quote(id) + mode)
		for index, label := range arrangementLabels {
			line := strconv.Itoa(index+1) + ". " + label + ":"
			if index == 3 {
				s.line(line)
				s.contactList(contacts, "   ", "   None.")
			} else {
				s.line(line + " " + quote(fields[index]))
			}
		}
		action, ok := s.ask(prompt)
		if !ok || action == "back" {
			return
		}
		if action == "id" && create {
			id = s.scalar("Permanent ID", id, nil)
			continue
		}
		if action == "apply" {
			duplicate := false
			if create {
				for _, item := range s.draft.Arrangements {
					if item.ID == id {
						duplicate = true
						break
					}
				}
			}
			if duplicate {
				s.line("Arrangement not applied.")
				s.line("Severity: \"error\" Code: \"guide.arrangement-exists\" Path: \"arrangements\"")
				continue
			}
			updated, problems := s.editor.SetArrangement(s.context, s.draft, arrangementFrom(id, fields, contacts))
			if len(problems) == 0 {
				s.draft = updated
				return
			}
			s.line("Arrangement not applied.")
			s.problems(problems)
			continue
		}
		index, valid := ordinal(action, 8)
		if !valid {
			s.invalid()
			continue
		}
		if index == 3 {
			contacts = s.helpers(contacts)
			continue
		}
		var choices func()
		switch index {
		case 0:
			choices = s.sections
		case 1:
			choices = s.kinds
		case 2:
			choices = s.contacts
		case 6:
			choices = s.statuses
		}
		fields[index] = s.scalar(arrangementLabels[index], fields[index], choices)
	}
}

func copyContacts(contacts []string) []string {
	if contacts == nil {
		return nil
	}
	return append(make([]string, 0, len(contacts)), contacts...)
}

func (s *session) helpers(held []string) []string {
	local := copyContacts(held)
	for s.active() {
		s.line("Assigned contacts — local list, not applied.")
		s.contactList(local, "", "No assigned contacts.")
		action, ok := s.ask("Helpers — not saved. Choose add, remove, clear, apply, or back:")
		if !ok || action == "back" {
			return held
		}
		switch action {
		case "apply":
			return local
		case "clear":
			local = nil
		case "add":
			s.contacts()
			value, valid := s.value("Enter the assigned contact ID (empty input retries):")
			if valid {
				local = append(local, value)
			}
		case "remove":
			index, selected := s.selection("Choose helper number or back:", len(local))
			if selected {
				local = append(local[:index], local[index+1:]...)
				if len(local) == 0 {
					local = nil
				}
			}
		default:
			s.invalid()
		}
	}
	return held
}

func (s *session) removal(index int) {
	s.line("Selected arrangement for removal:")
	s.arrangementBlock(index, s.draft.Arrangements[index])
	for s.active() {
		action, ok := s.ask("Removal is staged, not saved. Choose remove or back:")
		if !ok || action == "back" {
			return
		}
		if action != "remove" {
			s.invalid()
			continue
		}
		updated, problems := s.editor.RemoveArrangement(s.context, s.draft, s.draft.Arrangements[index].ID)
		if len(problems) != 0 {
			s.line("Arrangement not removed.")
			s.problems(problems)
		} else {
			s.draft = updated
		}
		return
	}
}

func (s *session) arrangementReview(label string, arrangements []model.Arrangement) {
	s.line(label)
	if len(arrangements) == 0 {
		s.line("None.")
	}
	for index, item := range arrangements {
		s.arrangementBlock(index, item)
	}
}

func (s *session) arrangementBlock(index int, item model.Arrangement) {
	s.line("Arrangement " + strconv.Itoa(index+1) + ":")
	s.line("ID: " + quote(item.ID))
	fields := arrangementValues(item)
	for index, label := range arrangementLabels {
		if index == 3 {
			s.line(label + ":")
			s.contactList(item.Contacts, "", "None.")
		} else {
			s.line(label + ": " + quote(fields[index]))
		}
	}
}

func (s *session) contactList(contacts []string, prefix, absent string) {
	if len(contacts) == 0 {
		s.line(absent)
	}
	for index, id := range contacts {
		s.line(prefix + strconv.Itoa(index+1) + ". " + quote(id))
	}
}
