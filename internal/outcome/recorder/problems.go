package recorder

import (
	"slices"

	contract "onboard-kit/contracts/outcome/recorder"
	"onboard-kit/internal/model"
)

func knownKind(kind model.OutcomeKind) bool {
	return kind == model.OutcomeVerified || kind == model.OutcomeFailed || kind == model.OutcomeDeferred
}

func problem(code, section, field, message string) model.Problem {
	return model.Problem{
		Code: "outcome." + code, Severity: model.SeverityError,
		Section: section, Message: message, Detail: map[string]string{"field": field},
	}
}

func proposalProblems(proposal contract.Proposal, on model.Timestamp) []model.Problem {
	var problems []model.Problem
	if !knownKind(proposal.Kind) {
		problems = append(problems, problem("kind-invalid", "", "outcome", "Choose verified, failed, or deferred for the new outcome."))
	}
	if !validTimestamp(string(on)) {
		problems = append(problems, problem("timestamp-invalid", "", "on", "Supply a valid UTC outcome timestamp."))
	}
	return problems
}

func heldKindProblems(outcomes []model.Outcome) []model.Problem {
	unknown := make(map[string]bool)
	for _, outcome := range outcomes {
		if !knownKind(outcome.Kind) {
			unknown[outcome.Section] = true
		}
	}
	sections := make([]string, 0, len(unknown))
	for section := range unknown {
		sections = append(sections, section)
	}
	slices.Sort(sections)
	var problems []model.Problem
	for _, section := range sections {
		problems = append(problems, problem("held-kind-invalid", section, "outcomes", "Repair the unknown recorded outcome kind before recording another outcome."))
	}
	return problems
}
