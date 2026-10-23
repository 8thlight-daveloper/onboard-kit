package assembler

import (
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

func bootstrapContext(section model.Section, parents map[string]string) []string {
	var blocks []string
	var ancestors []string
	for parent := section.Parent; parent != ""; parent = parents[parent] {
		ancestors = append(ancestors, parent)
	}
	slices.Reverse(ancestors)
	if len(ancestors) != 0 {
		blocks = append(blocks, "Within: "+quotedIdentifiers(ancestors, " / ")+".")
	}
	return append(blocks, requiresContext(section.Requires)...)
}

func requiresContext(identities []string) []string {
	var prerequisites []string
	seen := map[string]bool{}
	for _, identity := range identities {
		if !seen[identity] {
			seen[identity] = true
			prerequisites = append(prerequisites, identity)
		}
	}
	if len(prerequisites) != 0 {
		return []string{"Requires: " + quotedIdentifiers(prerequisites, ", ") + "."}
	}
	return nil
}

func quotedIdentifiers(identities []string, separator string) string {
	quoted := make([]string, len(identities))
	for index, identity := range identities {
		quoted[index] = "`" + identity + "`"
	}
	return strings.Join(quoted, separator)
}
