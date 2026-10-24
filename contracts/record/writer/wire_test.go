//go:build !contractstub

package writer_test

import (
	contract "onboard-kit/contracts/record/writer"
	implementation "onboard-kit/internal/record/writer"
)

func newWriter() contract.Writer {
	return implementation.New()
}
