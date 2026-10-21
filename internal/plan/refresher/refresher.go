package refresher

import "onboard-kit/internal/model"

type Refresher struct{}

func New() Refresher {
	return Refresher{}
}

func (Refresher) Refresh(base, theirs, ours model.Plan, origin model.Origin) (model.Plan, Report, []model.Problem) {
	if problem, refused := refusal(theirs, ours); refused {
		ours.Answers = copyAnswers(ours.Answers)
		return ours, Report{}, []model.Problem{problem}
	}
	theirQuestions := declaredQuestions(theirs)
	ourQuestions := declaredQuestions(ours)
	decisions := decideAll(base, theirs, ours, theirQuestions)

	refreshed := ours
	refreshed.Directory = theirs.Directory
	refreshed.Variables = theirs.Variables
	refreshed.Sections = theirs.Sections
	refreshed.UnknownFields = theirs.UnknownFields
	refreshed.Origin = model.Origin{PlanID: theirs.ID, Source: origin.Source, ContentHash: origin.ContentHash, RefreshedOn: origin.RefreshedOn}
	refreshed.Answers = answersAfter(ours.Answers, decisions)

	return refreshed, reportOf(theirs, ours, theirQuestions, ourQuestions, decisions), nil
}
