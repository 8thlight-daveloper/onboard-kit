package editor_test

import (
	contract "onboard-kit/contracts/preparation/editor"
	"onboard-kit/internal/model"
)

type stub struct{}

func newEditor() contract.Editor { return stub{} }

func (stub) SetRecipient(contract.Context, model.Preparation, model.PreparationRecipient) (model.Preparation, []model.Problem) {
	return model.Preparation{}, nil
}
func (stub) SetArrangement(contract.Context, model.Preparation, model.Arrangement) (model.Preparation, []model.Problem) {
	return model.Preparation{}, nil
}
func (stub) RemoveArrangement(contract.Context, model.Preparation, string) (model.Preparation, []model.Problem) {
	return model.Preparation{}, nil
}
