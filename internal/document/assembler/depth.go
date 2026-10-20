package assembler

import (
	"strings"

	"onboard-kit/internal/model"
)

const (
	rootHeadingLevel    = 2
	deepestHeadingLevel = 6
)

func parentsByID(sections []model.Section) map[string]string {
	parents := make(map[string]string, len(sections))
	for _, section := range sections {
		if _, seen := parents[section.ID]; !seen {
			parents[section.ID] = section.Parent
		}
	}
	return parents
}

func depthOf(section model.Section, parents map[string]string) int {
	depth := 0
	for parent := section.Parent; parent != ""; parent = parents[parent] {
		depth++
	}
	return depth
}

func headingMark(depth int) string {
	return strings.Repeat("#", min(rootHeadingLevel+depth, deepestHeadingLevel))
}
