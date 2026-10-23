package invite_test

import (
	"io"

	contract "onboard-kit/contracts/command/invite"
)

type emptyCommand struct{}

func (emptyCommand) Run([]string, io.Writer, io.Writer) int { return 0 }

func newCommand() contract.Command { return emptyCommand{} }
