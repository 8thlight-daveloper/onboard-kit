package materializer

import "onboard-kit/internal/model"

func copyStrings(source []string) []string {
	if source == nil {
		return nil
	}
	result := make([]string, len(source))
	copy(result, source)
	return result
}

func copyValue(source model.Value) model.Value {
	source.List = copyStrings(source.List)
	return source
}

func copyValues(source []model.Value) []model.Value {
	if source == nil {
		return nil
	}
	result := make([]model.Value, len(source))
	for i, value := range source {
		result[i] = copyValue(value)
	}
	return result
}

func copyCondition(source *model.Condition) *model.Condition {
	if source == nil {
		return nil
	}
	result := *source
	if source.All != nil {
		result.All = make([]model.Clause, len(source.All))
		for i, clause := range source.All {
			clause.Answer.Is = copyValue(clause.Answer.Is)
			clause.Answer.IsNot = copyValue(clause.Answer.IsNot)
			clause.Answer.IsOneOf = copyValues(clause.Answer.IsOneOf)
			clause.Answer.IsNotOneOf = copyValues(clause.Answer.IsNotOneOf)
			if clause.Answer.IsSettled != nil {
				settled := *clause.Answer.IsSettled
				clause.Answer.IsSettled = &settled
			}
			result.All[i] = clause
		}
	}
	return &result
}

func copyVariables(source []model.Variable) []model.Variable {
	if source == nil {
		return nil
	}
	result := make([]model.Variable, len(source))
	for i, variable := range source {
		variable.Default = copyValue(variable.Default)
		variable.Options = copyStrings(variable.Options)
		variable.AskIf = copyCondition(variable.AskIf)
		result[i] = variable
	}
	return result
}

func copyDirectory(source model.Directory) model.Directory {
	result := model.Directory{
		Contacts: make([]model.Contact, len(source.Contacts)),
		Roles:    make([]model.Role, len(source.Roles)),
		Teams:    make([]model.Team, len(source.Teams)),
	}
	copy(result.Contacts, source.Contacts)
	for i, role := range source.Roles {
		role.Holders = copyStrings(role.Holders)
		result.Roles[i] = role
	}
	for i, team := range source.Teams {
		if team.Holders != nil {
			holders := make(map[string][]string, len(team.Holders))
			for key, list := range team.Holders {
				holders[key] = copyStrings(list)
			}
			team.Holders = holders
		}
		result.Teams[i] = team
	}
	return result
}
