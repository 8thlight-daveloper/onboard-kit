package resolver

import "onboard-kit/internal/model"

const (
	deferredMarker   = "_deferred_"
	unansweredMarker = "_unanswered_"

	codeDeferred   = "gap.deferred"
	codeUnanswered = "gap.unanswered"
)

type resolution struct {
	text     string
	commands []string
	gapCode  string
}

func resolveVariable(variable model.Variable, answer model.Value, book book) resolution {
	for _, candidate := range []model.Value{answer, variable.Default} {
		if result, settled := settledResolution(variable, candidate, book); settled {
			return result
		}
		if candidate.Kind == model.ValueKindDeferred {
			return resolution{text: deferredMarker, gapCode: codeDeferred}
		}
	}
	if variable.Audience == model.AudienceOnboardee {
		return resolution{text: "<" + variable.Name + ">"}
	}
	return resolution{text: unansweredMarker, gapCode: codeUnanswered}
}

func settledResolution(variable model.Variable, value model.Value, book book) (resolution, bool) {
	if isDirectoryType(variable.Type) {
		text, found := book.written(variable.Type, value)
		return resolution{text: text}, found
	}
	if !isSettled(value) {
		return resolution{}, false
	}
	result := resolution{text: writtenValue(value)}
	if variable.Type == model.VariableTypeCommand && value.Kind == model.ValueKindList {
		result.commands = value.List
	}
	return result, true
}

func isDirectoryType(variableType model.VariableType) bool {
	switch variableType {
	case model.VariableTypeContact, model.VariableTypeRole, model.VariableTypeTeam:
		return true
	}
	return false
}
