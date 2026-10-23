package catalog_test

import (
	"io"

	contract "onboard-kit/contracts/command/catalog"
)

type stubCommand struct{}

func (stubCommand) Run([]string, io.Writer, io.Writer) int { return 0 }

func newCommand() contract.Command { return stubCommand{} }
