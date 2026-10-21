package validator

import "onboard-kit/internal/model"

func wrongKind(variable model.Variable, value model.Value) (model.Problem, bool) {
	expected := expectedKind(variable)
	if value.Kind == expected {
		return model.Problem{}, false
	}
	return model.Problem{
		Code:     "answer.wrong-kind",
		Severity: model.SeverityError,
		Message:  "This answer is " + kindName(value.Kind) + ", and the question takes " + kindName(expected) + ".",
		Detail:   map[string]string{"expected": kindName(expected), "found": kindName(value.Kind)},
	}, true
}

func expectedKind(variable model.Variable) model.ValueKind {
	switch {
	case typeOf(variable) == model.VariableTypeTeam:
		return model.ValueKindText
	case typeOf(variable) == model.VariableTypeCommand || variable.Multiple:
		return model.ValueKindList
	case typeOf(variable) == model.VariableTypeBoolean:
		return model.ValueKindFlag
	default:
		return model.ValueKindText
	}
}

func kindName(kind model.ValueKind) string {
	switch kind {
	case model.ValueKindFlag:
		return "flag"
	case model.ValueKindList:
		return "list"
	default:
		return "text"
	}
}
