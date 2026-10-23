package recorder

import (
	"slices"
	"strings"

	contract "onboard-kit/contracts/outcome/recorder"
	"onboard-kit/internal/model"
	"onboard-kit/internal/section/hasher"
	"onboard-kit/internal/section/walker"
)

type recorder struct{}

func New() contract.Recorder {
	return recorder{}
}

func (recorder) Record(plan model.Plan, record model.Record, proposal contract.Proposal, on model.Timestamp) (model.Record, []model.Problem) {
	result := copyRecord(record)
	if problems := proposalProblems(proposal, on); len(problems) > 0 {
		return result, problems
	}
	if problems := heldKindProblems(record.Outcomes); len(problems) > 0 {
		return result, problems
	}
	view, problems := walker.New().Walk(plan, record)
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return result, problems
		}
	}
	target := slices.IndexFunc(plan.Sections, func(section model.Section) bool {
		return section.ID == proposal.Section
	})
	if target < 0 {
		return result, append(problems, problem("section-unknown", "", "section", "Choose a section declared in this plan."))
	}
	row := -1
	for index, section := range view.Sections {
		if section.Section.ID == proposal.Section {
			row = index
			break
		}
	}
	if row < 0 {
		return result, append(problems, problem("section-inactive", proposal.Section, "section", "Choose an active section, or change its applicability explicitly first."))
	}
	if proposal.Kind == model.OutcomeVerified && len(view.Sections[row].WaitingOn) > 0 {
		for _, required := range view.Sections[row].WaitingOn {
			waiting := problem("prerequisite-unverified", proposal.Section, "requires", "Verify the required section before recording this section as verified.")
			waiting.Detail["required"] = required
			problems = append(problems, waiting)
		}
		return result, problems
	}
	outcome := model.Outcome{
		Section: proposal.Section, Kind: proposal.Kind, On: on,
		SectionHash: hasher.New().Hash(plan.Sections[target]),
		Note:        proposal.Note, Differently: proposal.Differently,
	}
	previous := slices.IndexFunc(result.Outcomes, func(held model.Outcome) bool {
		return held.Section == proposal.Section
	})
	if previous < 0 {
		result.Outcomes = append(result.Outcomes, outcome)
	} else {
		result.Outcomes[previous] = outcome
	}
	slices.SortFunc(result.Outcomes, func(left, right model.Outcome) int {
		return strings.Compare(left.Section, right.Section)
	})
	return result, problems
}
