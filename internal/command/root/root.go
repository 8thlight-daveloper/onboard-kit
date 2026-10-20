package root

import (
	"fmt"
	"io"

	"onboard-kit/internal/command/assemble"
	"onboard-kit/internal/command/create"
	"onboard-kit/internal/command/start"
	"onboard-kit/internal/command/stdio"
	"onboard-kit/internal/command/walk"
)

const exitUsage = 2

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: onboard-kit <command> [options]\ncommands: new, assemble, walk, start, stdio")
		return exitUsage
	}
	switch args[0] {
	case "new":
		return create.Run(args[1:], stdout, stderr)
	case "assemble":
		return assemble.Run(args[1:], stdout, stderr)
	case "walk":
		return walk.Run(args[1:], stdin, stdout, stderr)
	case "start":
		return start.Run(args[1:], stdin, stdout, stderr)
	case "stdio":
		return stdio.Run(args[1:], stdin, stdout, stderr)
	}
	fmt.Fprintf(stderr, "onboard-kit: unknown command %q\n", args[0])
	return exitUsage
}
