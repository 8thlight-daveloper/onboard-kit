package assembler

import (
	"strings"

	"onboard-kit/internal/model"
)

func sectionBlocks(section model.Section, declared []model.Variable, depth int) []string {
	return sectionBlocksWithContext(section, declared, depth, requiresContext(section.Requires))
}

func sectionBlocksWithContext(section model.Section, declared []model.Variable, depth int, context []string) []string {
	text := prose{blanks: onboardingBlanks(declared)}
	var blocks []string
	if text.present(section.Title) {
		blocks = append(blocks, headingMark(depth)+" "+text.of(section.Title))
	}
	blocks = append(blocks, context...)
	if text.present(section.Summary) {
		blocks = append(blocks, text.of(section.Summary))
	}
	if len(section.Steps) > 0 {
		blocks = append(blocks, stepList(section, text))
	}
	if text.present(section.Body) {
		blocks = append(blocks, text.of(section.Body))
	}
	if text.present(section.Verification) {
		blocks = append(blocks, "Done when: "+text.of(section.Verification))
	}
	if footer := sectionFooter(section, text); footer != "" {
		blocks = append(blocks, footer)
	}
	return blocks
}

func stepList(section model.Section, text prose) string {
	items := make([]string, len(section.Steps))
	for index, step := range section.Steps {
		items[index] = stepItem(step, section.Timing, text)
	}
	return strings.Join(items, lineBreak)
}
