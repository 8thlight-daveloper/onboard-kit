package root

import (
	"fmt"
	"io"
	"time"

	sessioneditorapi "onboard-kit/contracts/session/editor"
	catalogcreator "onboard-kit/internal/catalog/creator"
	"onboard-kit/internal/catalog/editguide"
	"onboard-kit/internal/command/assemble"
	"onboard-kit/internal/command/catalog"
	"onboard-kit/internal/command/create"
	"onboard-kit/internal/command/invite"
	"onboard-kit/internal/command/prepare"
	"onboard-kit/internal/command/stdio"
	"onboard-kit/internal/command/walk"
	sessioneditor "onboard-kit/internal/session/editor"
)

const exitUsage = 2

var version = "development"

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: onboard-kit <command> [options]\ncommands: new, assemble, walk, start, prepare, invite, catalog, stdio")
		return exitUsage
	}
	switch args[0] {
	case "--version":
		fmt.Fprintf(stdout, "onboard-kit %s\n", version)
		return 0
	case "new":
		return create.Run(args[1:], stdout, stderr)
	case "assemble":
		return assemble.Run(args[1:], stdout, stderr)
	case "walk":
		return walk.Run(args[1:], version, stdin, stdout, stderr)
	case "start":
		return startCommand(args[1:], stdin, stdout, stderr)
	case "prepare":
		return prepare.New().Run(args[1:], stdout, stderr)
	case "invite":
		return invite.New().Run(args[1:], stdout, stderr)
	case "catalog":
		if len(args) > 1 && args[1] == "create" {
			return catalogcreator.New().Run(args[2:], stdin, stdout, stderr)
		}
		if len(args) > 1 && args[1] == "edit" {
			return editguide.New().Run(args[2:], stdin, stdout, stderr)
		}
		return catalog.New().Run(args[1:], stdout, stderr)
	case "stdio":
		return stdio.Run(args[1:], stdin, stdout, stderr)
	case "session":
		if len(args) > 1 && args[1] == "edit" {
			return sessioneditor.New(sessionClock{}).Run(args[2:], sessioneditorapi.Streams{Input: stdin, Output: stdout, Error: stderr}).Exit
		}
		fmt.Fprintln(stderr, "usage: onboard-kit session edit --root ABS --home ABS --session KEY")
		return exitUsage
	}
	fmt.Fprintf(stderr, "onboard-kit: unknown command %q\n", args[0])
	return exitUsage
}

type sessionClock struct{}

func (sessionClock) Now() string { return time.Now().UTC().Format(time.RFC3339Nano) }
