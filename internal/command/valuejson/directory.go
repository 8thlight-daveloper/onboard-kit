package valuejson

import (
	"bytes"
	"encoding/json"

	"onboard-kit/internal/model"
)

type contactEntry struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Slack   string `json:"slack"`
	SlackID string `json:"slackId"`
}

type roleEntry struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Holders []string `json:"holders"`
}

type teamEntry struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	Contact string              `json:"contact"`
	Holders map[string][]string `json:"holders"`
}

func DecodeContact(data json.RawMessage) (model.Contact, error) {
	var entry contactEntry
	if err := decodeStrictly(data, &entry); err != nil {
		return model.Contact{}, err
	}
	return model.Contact(entry), nil
}

func DecodeRole(data json.RawMessage) (model.Role, error) {
	var entry roleEntry
	if err := decodeStrictly(data, &entry); err != nil {
		return model.Role{}, err
	}
	return model.Role(entry), nil
}

func DecodeTeam(data json.RawMessage) (model.Team, error) {
	var entry teamEntry
	if err := decodeStrictly(data, &entry); err != nil {
		return model.Team{}, err
	}
	return model.Team(entry), nil
}

func decodeStrictly(data json.RawMessage, into any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(into)
}

func EncodeDirectory(directory model.Directory) map[string]any {
	contacts := make([]any, 0, len(directory.Contacts))
	for _, contact := range directory.Contacts {
		contacts = append(contacts, contactEntry(contact))
	}
	roles := make([]any, 0, len(directory.Roles))
	for _, role := range directory.Roles {
		roles = append(roles, roleEntry{ID: role.ID, Name: role.Name, Holders: nonNil(role.Holders)})
	}
	teams := make([]any, 0, len(directory.Teams))
	for _, team := range directory.Teams {
		held := map[string][]string{}
		for role, holders := range team.Holders {
			held[role] = nonNil(holders)
		}
		teams = append(teams, teamEntry{ID: team.ID, Name: team.Name, Holders: held})
	}
	return map[string]any{"contacts": contacts, "roles": roles, "teams": teams}
}

func EncodeCleared(cleared model.Cleared) map[string]any {
	return map[string]any{"answers": nonNil(cleared.Answers), "holders": nonNil(cleared.Holders)}
}

func nonNil(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}
