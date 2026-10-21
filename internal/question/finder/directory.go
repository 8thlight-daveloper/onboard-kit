package finder

import "onboard-kit/internal/model"

func isDirectoryType(variableType model.VariableType) bool {
	switch variableType {
	case model.VariableTypeContact, model.VariableTypeRole, model.VariableTypeTeam:
		return true
	}
	return false
}

func directoryChoices(directory model.Directory, variableType model.VariableType) []model.Choice {
	choices := []model.Choice{}
	switch variableType {
	case model.VariableTypeContact:
		for _, contact := range directory.Contacts {
			choices = append(choices, model.Choice{Value: contact.ID, Label: contact.Name})
		}
	case model.VariableTypeRole:
		for _, role := range directory.Roles {
			choices = append(choices, model.Choice{Value: role.ID, Label: role.Name})
		}
	case model.VariableTypeTeam:
		for _, team := range directory.Teams {
			choices = append(choices, model.Choice{Value: team.ID, Label: team.Name})
		}
	}
	return choices
}

func hasChoice(choices []model.Choice, id string) bool {
	for _, choice := range choices {
		if choice.Value == id {
			return true
		}
	}
	return false
}
