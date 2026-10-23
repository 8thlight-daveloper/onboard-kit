package assembler

import (
	"strconv"

	contract "onboard-kit/contracts/firststeps/assembler"
	preparationcontract "onboard-kit/contracts/preparation/validator"
	documentassembler "onboard-kit/internal/document/assembler"
	"onboard-kit/internal/model"
	"onboard-kit/internal/preparation/locator"
	"onboard-kit/internal/preparation/validator"
)

type assembler struct{}

func New() contract.Assembler {
	return assembler{}
}

func (assembler) Assemble(context contract.Context, plan model.Plan, preparation model.Preparation) (string, []model.Problem) {
	if problems := validator.New().Validate(preparationContext(context.Account, plan), preparation); len(problems) > 0 {
		return "", problems
	}
	directory := newDirectory(plan.Directory, preparation.Recipient.Team)
	if problems := directory.validateHelp(context.GeneralHelp); len(problems) > 0 {
		return "", problems
	}
	paths, problems := locator.New().Locate(plan.ResultsPath, context.Account)
	if len(problems) > 0 {
		return "", problems
	}
	fragment, problems := documentassembler.New().AssembleBootstrap(plan)
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return "", problems
		}
	}
	return guide(context, plan, preparation, directory, paths, fragment), problems
}

func preparationContext(account string, plan model.Plan) preparationcontract.Context {
	context := preparationcontract.Context{Account: account, PlanID: plan.ID}
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

func (directory directory) validateHelp(references []contract.HelpReference) []model.Problem {
	var problems []model.Problem
	seen := make(map[contract.HelpReference]bool)
	for index, reference := range references {
		if seen[reference] {
			continue
		}
		seen[reference] = true
		if directory.has(reference) {
			continue
		}
		problems = append(problems, model.Problem{
			Code: "firststeps.help-reference", Severity: model.SeverityError,
			Message: "Choose an existing contact or role for this help item, or remove it.",
			Detail:  map[string]string{"field": "generalHelp", "item": strconv.Itoa(index)},
		})
	}
	return problems
}
