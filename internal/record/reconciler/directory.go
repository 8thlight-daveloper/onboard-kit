package reconciler

import (
	"slices"

	"onboard-kit/internal/model"
)

func isDirectoryType(variableType model.VariableType) bool {
	switch variableType {
	case model.VariableTypeContact, model.VariableTypeRole, model.VariableTypeTeam:
		return true
	default:
		return false
	}
}

func directoryIDs(directory model.Directory, variableType model.VariableType) []string {
	var ids []string
	switch variableType {
	case model.VariableTypeContact:
		for _, contact := range directory.Contacts {
			ids = append(ids, contact.ID)
		}
	case model.VariableTypeRole:
		for _, role := range directory.Roles {
			ids = append(ids, role.ID)
		}
	case model.VariableTypeTeam:
		for _, team := range directory.Teams {
			ids = append(ids, team.ID)
		}
	}
	return ids
}

func directoryKind(variable model.Variable) model.ValueKind {
	if variable.Multiple && variable.Type != model.VariableTypeTeam {
		return model.ValueKindList
	}
	return model.ValueKindText
}

func checkedDirectoryAnswer(value model.Value, variable model.Variable, directory model.Directory) (model.Value, listing) {
	if value.Kind != directoryKind(variable) {
		return model.Value{}, listingReview
	}
	ids := directoryIDs(directory, variable.Type)
	if value.Kind == model.ValueKindText {
		return keptIfValid(value, slices.Contains(ids, value.Text))
	}
	return checkedIDList(value, ids)
}

func checkedIDList(value model.Value, ids []string) (model.Value, listing) {
	remaining := slices.DeleteFunc(slices.Clone(value.List), func(item string) bool {
		return item != "" && !slices.Contains(ids, item)
	})
	switch {
	case len(remaining) == len(value.List):
		return value, listingKept
	case slices.ContainsFunc(remaining, isNonEmpty):
		return model.Value{Kind: model.ValueKindList, List: remaining}, listingReview
	default:
		return model.Value{}, listingReview
	}
}
