package tui

import (
	editorcontract "onboard-kit/contracts/preparation/editor"
	reviewercontract "onboard-kit/contracts/preparation/reviewer"
	"onboard-kit/internal/model"
	"onboard-kit/internal/preparation/editor"
	"onboard-kit/internal/preparation/encoder"
	"onboard-kit/internal/preparation/reviewer"
)

const preparationByteLimit = 1 << 20

func editorContext(account string, plan model.Plan) editorcontract.Context {
	context := editorcontract.Context{Account: account, PlanID: plan.ID}
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

func (change preparationChange) applied(loaded PreparationLoad, current model.Preparation) (model.Preparation, []model.Problem) {
	context := editorContext(loaded.Account, loaded.Plan)
	switch change.kind {
	case changeRecipient:
		return editor.New().SetRecipient(context, current, change.recipient)
	case changeRemoval:
		return editor.New().RemoveArrangement(context, current, change.arrangement.ID)
	}
	return editor.New().SetArrangement(context, current, change.arrangement)
}

func reviewOf(loaded PreparationLoad, proposed model.Preparation) reviewercontract.Review {
	return reviewer.New().Review(reviewercontract.Context{Account: loaded.Account}, loaded.Plan, proposed)
}

func encoded(proposed model.Preparation) ([]byte, []model.Problem) {
	return encoder.New().Encode(proposed)
}
