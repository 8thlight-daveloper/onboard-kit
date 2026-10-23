package guide_test

import (
	"io"

	contract "onboard-kit/contracts/preparation/guide"
)

type stubCommand struct{}

func (stubCommand) Run([]string, io.Reader, io.Writer, io.Writer) int { return 0 }

func newCommand() contract.Command { return stubCommand{} }
