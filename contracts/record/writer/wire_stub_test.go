//go:build contractstub

package writer_test

import (
	contract "onboard-kit/contracts/record/writer"
	"onboard-kit/internal/model"
)

type emptyWriter struct{}

func newWriter() contract.Writer {
	return emptyWriter{}
}

func (emptyWriter) Write(model.Plan, model.Record, contract.Intent) (model.Record, []model.Problem) {
	return model.Record{}, nil
}
