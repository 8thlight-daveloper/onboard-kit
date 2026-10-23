package walker

import (
	"onboard-kit/internal/model"
	"onboard-kit/internal/question/finder"
	"onboard-kit/internal/section/hasher"
	"onboard-kit/internal/section/resolver"
	"onboard-kit/internal/section/scheduler"
)

type Walker struct{}

func New() Walker {
	return Walker{}
}

func (Walker) Walk(plan model.Plan, record model.Record) (View, []model.Problem) {
	sections, problems := scheduler.New().Schedule(plan)
	if len(problems) != 0 {
		return View{}, problems
	}
	if problems = duplicateOutcomes(record.Outcomes); len(problems) != 0 {
		return View{}, problems
	}
	hashes := make(map[string]string, len(plan.Sections))
	for _, section := range plan.Sections {
		hashes[section.ID] = hasher.New().Hash(section)
	}
	outcomes := make(map[string]model.Outcome, len(record.Outcomes))
	for _, outcome := range record.Outcomes {
		outcomes[outcome.Section] = outcome
	}
	questions := finder.New()
	view := View{
		Outcomes:  outcomeViews(record.Outcomes, hashes),
		Questions: questions.Questions(plan, ""),
		Blockers:  questions.Blockers(plan, ""),
	}
	team := teamIdentifier(plan)
	for _, section := range sections {
		resolved, gaps := resolver.New().Resolve(section, plan.Answers, plan.Directory, team)
		escalation, supportGaps := resolveEscalation(section, plan, team)
		view.Sections = append(view.Sections, SectionView{
			Section:     resolved,
			CurrentHash: hashes[section.ID],
			WaitingOn:   waitingOn(section.Requires, outcomes),
			Escalation:  escalation,
		})
		problems = append(problems, sectionProblems(section, append(gaps, supportGaps...))...)
	}
	return view, problems
}
