//go:build contractstub

package editor_test

import contract "onboard-kit/contracts/session/editor"

type emptyDriver struct{}

func newDriver(contract.Clock) contract.Driver {
	return emptyDriver{}
}

func (emptyDriver) Run([]string, contract.Streams) contract.Result {
	return contract.Result{}
}
