package resolver

import (
	"strconv"
	"strings"

	"onboard-kit/internal/model"
)

const (
	personSeparator = ", then "
	listSeparator   = " "
)

func writtenValue(value model.Value, variableType model.VariableType) string {
	switch value.Kind {
	case model.ValueKindFlag:
		return strconv.FormatBool(value.Flag)
	case model.ValueKindList:
		return strings.Join(value.List, listSeparatorFor(variableType))
	default:
		return value.Text
	}
}

func listSeparatorFor(variableType model.VariableType) string {
	if variableType == model.VariableTypePerson {
		return personSeparator
	}
	return listSeparator
}
