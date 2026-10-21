package resolver

import (
	"strings"

	"onboard-kit/internal/model"
)

const directorySeparator = "; "

type book struct {
	directory model.Directory
	team      string
}

func newBook(directory model.Directory, team string) book {
	return book{directory: directory, team: team}
}

func (b book) written(variableType model.VariableType, value model.Value) (string, bool) {
	var texts []string
	for _, id := range answeredIDs(value) {
		if text, found := b.writtenID(variableType, id); found {
			texts = append(texts, text)
		}
	}
	return strings.Join(texts, directorySeparator), len(texts) > 0
}

func answeredIDs(value model.Value) []string {
	switch value.Kind {
	case model.ValueKindText:
		return []string{value.Text}
	case model.ValueKindList:
		return value.List
	default:
		return nil
	}
}

func (b book) writtenID(variableType model.VariableType, id string) (string, bool) {
	switch variableType {
	case model.VariableTypeContact:
		return b.writtenContact(id)
	case model.VariableTypeRole:
		return b.writtenRole(id)
	default:
		return b.writtenTeam(id)
	}
}

func (b book) writtenContact(id string) (string, bool) {
	for _, contact := range b.directory.Contacts {
		if contact.ID == id {
			return contactText(contact), true
		}
	}
	return "", false
}

func contactText(contact model.Contact) string {
	switch {
	case contact.Slack != "":
		return contact.Name + ", " + contact.Slack
	case contact.Email != "":
		return contact.Name + ", " + contact.Email
	default:
		return contact.Name
	}
}

func (b book) writtenTeam(id string) (string, bool) {
	for _, team := range b.directory.Teams {
		if team.ID == id {
			return team.Name, true
		}
	}
	return "", false
}
