package assembler

import (
	"strings"

	"onboard-kit/internal/model"
	"onboard-kit/internal/section/scheduler"
)

func (a Assembler) AssembleBootstrap(plan model.Plan) (string, []model.Problem) {
	scheduled, structuralProblems := scheduler.New().Schedule(plan)
	if len(structuralProblems) != 0 {
		return "", structuralProblems
	}
	context := planContext{answers: plan.Answers, directory: plan.Directory, team: teamIdentifier(plan)}
	var sections []filledSection
	var problems []model.Problem
	for _, active := range scheduled {
		if !active.Bootstrap {
			continue
		}
		filled, sectionProblems := a.fillSection(active, context)
		sections = append(sections, filledSection{
			filled: filled, declared: active.Variables, problems: sectionProblems,
		})
		problems = append(problems, sectionProblems...)
	}
	if len(sections) == 0 {
		return "", nil
	}
	parents := parentsByID(plan.Sections)
	blocks := []string{gapTableAtLevel(sections, rootHeadingLevel+1)}
	for _, section := range sections {
		contextBlocks := bootstrapContext(section.filled, parents)
		blocks = append(blocks, sectionBlocksWithContext(section.filled, section.declared, 1, contextBlocks)...)
	}
	return strings.Join(blocks, blockSeparator) + lineBreak, problems
}
