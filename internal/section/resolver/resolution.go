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

func resolveVariable(variable model.Variable, answer model.Value) resolution {
	value, kind := chooseValue(variable, answer)
	switch kind {
	case valueSettled:
		return settledResolution(variable, value)
	case valueDeferred:
		return resolution{text: deferredMarker, gapCode: codeDeferred}
	}
	if variable.Audience == model.AudienceOnboardee {
		return resolution{text: "<" + variable.Name + ">"}
	}
	return resolution{text: unansweredMarker, gapCode: codeUnanswered}
}

func settledResolution(variable model.Variable, value model.Value) resolution {
	result := resolution{text: writtenValue(value, variable.Type)}
	if variable.Type == model.VariableTypeCommand && value.Kind == model.ValueKindList {
		result.commands = value.List
	}
	return result
}

type valueKind int

const (
	valueMissing valueKind = iota
	valueSettled
	valueDeferred
)

func chooseValue(variable model.Variable, answer model.Value) (model.Value, valueKind) {
	if isSettled(answer) {
		return answer, valueSettled
	}
	if answer.Kind == model.ValueKindDeferred {
		return answer, valueDeferred
	}
	if isSettled(variable.Default) {
		return variable.Default, valueSettled
	}
	if variable.Default.Kind == model.ValueKindDeferred {
		return variable.Default, valueDeferred
	}
	return model.Value{}, valueMissing
}
