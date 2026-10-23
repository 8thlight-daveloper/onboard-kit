package reader_test

import (
	contract "onboard-kit/contracts/catalog/reader"
	implementation "onboard-kit/internal/catalog/reader"
)

func newReader() contract.Reader { return implementation.New() }
