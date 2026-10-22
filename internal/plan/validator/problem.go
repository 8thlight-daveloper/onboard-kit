package validator

import "onboard-kit/internal/model"

func annotate(plan model.Plan, target question, problems []model.Problem) []model.Problem {
	section := sectionOf(plan, target)
	for position := range problems {
		problems[position].Question = target.id
		problems[position].Section = section
	}
	return problems
}

func unknownQuestion(plan model.Plan, target question, severity model.Severity) model.Problem {
	return annotate(plan, target, []model.Problem{{
		Code:     "answer.unknown-question",
		Severity: severity,
		Message:  "No variable declares this question.",
	}})[0]
}

func inactiveQuestion(plan model.Plan, target question) model.Problem {
	return annotate(plan, target, []model.Problem{{
		Code:     "answer.inactive-question",
		Severity: model.SeverityError,
		Message:  "This question is not asked in the plan as it stands.",
	}})[0]
}

func ambiguousQuestion(plan model.Plan, target question) model.Problem {
	return annotate(plan, target, []model.Problem{{
		Code:     "answer.ambiguous-question",
		Severity: model.SeverityError,
		Message:  "More than one variable declares this question. Give each variable in its scope a unique name.",
	}})[0]
}
