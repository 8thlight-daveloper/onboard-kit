package assembler

import (
	"maps"
	"slices"
	"strings"
)

func environmentPart(environment map[string]string) string {
	var assignments []string
	for _, name := range slices.Sorted(maps.Keys(environment)) {
		assignments = append(assignments, "`"+name+"="+environment[name]+"`")
	}
	return "Env: " + strings.Join(assignments, ", ") + "."
}
