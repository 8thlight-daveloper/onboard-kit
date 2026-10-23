package reviewer

import (
	"maps"
	"slices"

	firststepscontract "onboard-kit/contracts/firststeps/assembler"
	contract "onboard-kit/contracts/preparation/reviewer"
	validationcontract "onboard-kit/contracts/preparation/validator"
	firststepsassembler "onboard-kit/internal/firststeps/assembler"
	"onboard-kit/internal/model"
	"onboard-kit/internal/preparation/validator"
	"onboard-kit/internal/question/finder"
)

type reviewer struct {
	preparationValidator validationcontract.Validator
	questionFinder       finder.Finder
	firststepsAssembler  firststepscontract.Assembler
}

func New() contract.Reviewer {
	return reviewer{
		preparationValidator: validator.New(),
		questionFinder:       finder.New(),
		firststepsAssembler:  firststepsassembler.New(),
	}
}

func (reviewer reviewer) Review(context contract.Context, originalPlan model.Plan, proposedPreparation model.Preparation) contract.Review {
	problems := reviewer.preparationValidator.Validate(validationContext(context.Account, originalPlan), proposedPreparation)
	result := contract.Review{PreparationProblems: copyProblems(problems)}
	if len(problems) > 0 {
		return result
	}

	result.OnboarderBlockers = slices.Clone(reviewer.questionFinder.Blockers(originalPlan, model.AudienceOnboarder))
	text, problems := reviewer.firststepsAssembler.Assemble(firststepsContext(context), originalPlan, proposedPreparation)
	result.FirstSteps = text
	result.FirstStepsProblems = copyProblems(problems)
	return result
}

func validationContext(account string, plan model.Plan) validationcontract.Context {
	context := validationcontract.Context{Account: account, PlanID: plan.ID}
	for _, contact := range plan.Directory.Contacts {
		context.Contacts = append(context.Contacts, contact.ID)
	}
	for _, team := range plan.Directory.Teams {
		context.Teams = append(context.Teams, team.ID)
	}
	for _, section := range plan.Sections {
		context.Sections = append(context.Sections, section.ID)
	}
	return context
}

func firststepsContext(context contract.Context) firststepscontract.Context {
	result := firststepscontract.Context{Account: context.Account}
	for _, reference := range context.GeneralHelp {
		result.GeneralHelp = append(result.GeneralHelp, firststepscontract.HelpReference{
			Kind: reference.Kind,
			ID:   reference.ID,
		})
	}
	return result
}

func copyProblems(problems []model.Problem) []model.Problem {
	result := slices.Clone(problems)
	for index := range result {
		result[index].Detail = maps.Clone(problems[index].Detail)
	}
	return result
}
