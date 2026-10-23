package editor

import (
	"slices"
	"strconv"
	"strings"

	contract "onboard-kit/contracts/preparation/editor"
	validation "onboard-kit/contracts/preparation/validator"
	"onboard-kit/internal/model"
	"onboard-kit/internal/preparation/validator"
)

type editor struct{}

func New() contract.Editor {
	return editor{}
}

func (editor) SetRecipient(context contract.Context, current model.Preparation, recipient model.PreparationRecipient) (model.Preparation, []model.Problem) {
	if problems := guard(context, current); len(problems) > 0 {
		return current, problems
	}
	projection := identity(current)
	projection.Recipient = recipient
	if problems := validate(context, projection); len(problems) > 0 {
		return current, problems
	}
	result := current
	result.Recipient = recipient
	return result, nil
}

func (editor) SetArrangement(context contract.Context, current model.Preparation, arrangement model.Arrangement) (model.Preparation, []model.Problem) {
	if problems := guard(context, current); len(problems) > 0 {
		return current, problems
	}
	index := arrangementIndex(current.Arrangements, arrangement.ID)
	if index < 0 {
		index = len(current.Arrangements)
	}
	projection := identity(current)
	projection.Arrangements = []model.Arrangement{arrangement}
	if problems := validate(context, projection); len(problems) > 0 {
		for position := range problems {
			path := problems[position].Detail["path"]
			if suffix, found := strings.CutPrefix(path, "arrangements[0]"); found {
				problems[position].Detail["path"] = "arrangements[" + strconv.Itoa(index) + "]" + suffix
			}
		}
		return current, problems
	}
	result := current
	result.Arrangements = copyArrangements(current.Arrangements)
	arrangement.Contacts = slices.Clone(arrangement.Contacts)
	if index == len(result.Arrangements) {
		result.Arrangements = append(result.Arrangements, arrangement)
	} else {
		result.Arrangements[index] = arrangement
	}
	return result, nil
}

func (editor) RemoveArrangement(context contract.Context, current model.Preparation, id string) (model.Preparation, []model.Problem) {
	if problems := guard(context, current); len(problems) > 0 {
		return current, problems
	}
	index := arrangementIndex(current.Arrangements, id)
	if index < 0 {
		return current, []model.Problem{problem("preparation.unknown-arrangement", "arrangements", "Choose an existing arrangement to remove.")}
	}
	result := current
	result.Arrangements = nil
	if len(current.Arrangements) > 1 {
		result.Arrangements = copyArrangements(current.Arrangements)
		result.Arrangements = slices.Delete(result.Arrangements, index, index+1)
	}
	return result, nil
}

func identity(current model.Preparation) model.Preparation {
	return model.Preparation{
		SchemaVersion: current.SchemaVersion,
		Kind:          current.Kind,
		Account:       current.Account,
		PlanID:        current.PlanID,
	}
}

func validate(context contract.Context, preparation model.Preparation) []model.Problem {
	return validator.New().Validate(validation.Context{
		Account:  context.Account,
		PlanID:   context.PlanID,
		Contacts: context.Contacts,
		Teams:    context.Teams,
		Sections: context.Sections,
	}, preparation)
}

func guard(context contract.Context, current model.Preparation) []model.Problem {
	if problems := validate(context, identity(current)); len(problems) > 0 {
		return problems
	}
	seen := make(map[string]bool, len(current.Arrangements))
	var problems []model.Problem
	for index, arrangement := range current.Arrangements {
		if seen[arrangement.ID] {
			path := "arrangements[" + strconv.Itoa(index) + "].id"
			problems = append(problems, problem("preparation.ambiguous-arrangement", path, "Give each arrangement a distinct permanent identifier before editing."))
		}
		seen[arrangement.ID] = true
	}
	return problems
}

func arrangementIndex(arrangements []model.Arrangement, id string) int {
	for index, arrangement := range arrangements {
		if arrangement.ID == id {
			return index
		}
	}
	return -1
}

func copyArrangements(arrangements []model.Arrangement) []model.Arrangement {
	result := slices.Clone(arrangements)
	for index := range result {
		result[index].Contacts = slices.Clone(result[index].Contacts)
	}
	return result
}

func problem(code, path, message string) model.Problem {
	return model.Problem{
		Severity: model.SeverityError,
		Code:     code,
		Message:  message,
		Detail:   map[string]string{"path": path},
	}
}
