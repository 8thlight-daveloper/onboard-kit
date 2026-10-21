package resolver

import (
	"strconv"
	"strings"

	"onboard-kit/internal/model"
)

const listSeparator = " "

func writtenValue(value model.Value) string {
	switch value.Kind {
	case model.ValueKindFlag:
		return strconv.FormatBool(value.Flag)
	case model.ValueKindList:
		return strings.Join(value.List, listSeparator)
	default:
		return value.Text
	}
}
