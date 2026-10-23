package editor

import "onboard-kit/internal/model"

type Context struct {
	Account  string
	PlanID   string
	Contacts []string
	Teams    []string
	Sections []string
}

type Editor interface {
	SetRecipient(Context, model.Preparation, model.PreparationRecipient) (model.Preparation, []model.Problem)
	SetArrangement(Context, model.Preparation, model.Arrangement) (model.Preparation, []model.Problem)
	RemoveArrangement(Context, model.Preparation, string) (model.Preparation, []model.Problem)
}
