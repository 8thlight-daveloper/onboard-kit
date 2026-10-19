package finder

import (
	"slices"
	"strconv"
	"strings"

	"onboard-kit/internal/model"
)

func groupKey(question model.Question) string {
	options := slices.Clone(question.Options)
	slices.Sort(options)
	options = slices.Compact(options)
	parts := []string{strconv.Quote(question.Name), strconv.Quote(string(question.Type)), strconv.FormatBool(question.Multiple)}
	for _, option := range options {
		parts = append(parts, strconv.Quote(option))
	}
	return strings.Join(parts, ",")
}
