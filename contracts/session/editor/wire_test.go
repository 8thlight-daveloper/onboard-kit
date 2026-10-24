//go:build !contractstub

package editor_test

import (
	contract "onboard-kit/contracts/session/editor"
	implementation "onboard-kit/internal/session/editor"
)

func newDriver(clock contract.Clock) contract.Driver {
	return implementation.New(clock)
}
