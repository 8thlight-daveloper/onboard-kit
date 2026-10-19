package assembler

import (
	"strings"

	"onboard-kit/internal/model"
	"onboard-kit/internal/section/resolver"
	"onboard-kit/internal/section/selector"
)

type Assembler struct {
	selector selector.Selector
	resolver resolver.Resolver
}

func New() Assembler {
	return Assembler{selector: selector.New(), resolver: resolver.New()}
}

func (a Assembler) Assemble(plan model.Plan) (string, []model.Problem) {
	filled, problems := a.fillActiveSections(plan)
	blocks := []string{
		titleLine(plan.Answers),
		reviewLine(plan.Provenance),
		gapTable(filled),
	}
	for _, section := range filled {
		blocks = append(blocks, sectionBlocks(section.filled, section.declared)...)
	}
	return strings.Join(blocks, blockSeparator) + lineBreak, problems
}
