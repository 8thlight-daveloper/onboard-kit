package recorder_test

import (
	contract "onboard-kit/contracts/outcome/recorder"
	"onboard-kit/internal/model"
)

type stubRecorder struct{}

func (stubRecorder) Record(model.Plan, model.Record, contract.Proposal, model.Timestamp) (model.Record, []model.Problem) {
	return model.Record{}, nil
}

func newRecorder() contract.Recorder {
	return stubRecorder{}
}
