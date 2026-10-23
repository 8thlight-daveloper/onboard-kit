package walker_test

import (
	contract "onboard-kit/contracts/section/walker"
	"onboard-kit/internal/model"
)

type stubWalker struct{}

func (stubWalker) Walk(model.Plan, model.Record) (contract.View, []model.Problem) {
	return contract.View{}, nil
}

func newWalker() contract.Walker {
	return stubWalker{}
}
