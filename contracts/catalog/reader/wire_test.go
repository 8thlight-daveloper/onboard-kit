package reader_test

import (
	contract "onboard-kit/contracts/catalog/reader"
	"onboard-kit/internal/model"
)

type stubReader struct{}

func (stubReader) List() []contract.Summary { return nil }

func (stubReader) Read(string) (contract.Entry, []model.Problem) { return contract.Entry{}, nil }

func newReader() contract.Reader { return stubReader{} }
