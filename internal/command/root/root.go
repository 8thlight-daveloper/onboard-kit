package root

import (
	"fmt"
	"io"

	"onboard-kit/internal/command/assemble"
	"onboard-kit/internal/command/create"
)

const exitUsage = 2

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: onboard-kit <command> [options]\ncommands: new, assemble")
		return exitUsage
	}
	switch args[0] {
	case "new":
		return create.Run(args[1:], stdout, stderr)
	case "assemble":
		return assemble.Run(args[1:], stdout, stderr)
	}
	fmt.Fprintf(stderr, "onboard-kit: unknown command %q\n", args[0])
	return exitUsage
}
