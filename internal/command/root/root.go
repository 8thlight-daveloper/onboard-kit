package root

import (
	"fmt"
	"io"
	"os"
	"time"

	sessioneditorapi "onboard-kit/contracts/session/editor"
	catalogcreator "onboard-kit/internal/catalog/creator"
	"onboard-kit/internal/catalog/editguide"
	"onboard-kit/internal/command/assemble"
	"onboard-kit/internal/command/catalog"
	"onboard-kit/internal/command/frontdoor"
	"onboard-kit/internal/command/invite"
	"onboard-kit/internal/command/prepare"
	"onboard-kit/internal/command/stdio"
	"onboard-kit/internal/command/walk"
	sessioneditor "onboard-kit/internal/session/editor"
)

const (
	exitUsage = 2
	usageText = "usage: onboard-kit [FILE] | <command> [options]\ncommands: new, open, assemble, walk, start, prepare, invite, catalog, stdio"
)

var version = "development"

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		if !frontdoor.Terminal(stdin, stdout) {
			fmt.Fprintln(stderr, usageText)
			fmt.Fprintln(stderr, "The interactive tool needs a terminal. "+frontdoor.ScriptingCommands)
			return exitUsage
		}
		return frontdoor.Run(version, stdin, stdout, stderr)
	}
	switch args[0] {
	case "--version":
		fmt.Fprintf(stdout, "onboard-kit %s\n", version)
		return 0
	case "new":
		return frontdoor.New(args[1:], version, stdin, stdout, stderr)
	case "open":
		return frontdoor.Open(args[1:], version, stdin, stdout, stderr)
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
	if names, ok := namesAFile(args); ok {
		return frontdoor.Open(names, version, stdin, stdout, stderr)
	}
	fmt.Fprintf(stderr, "onboard-kit: unknown command %q\n", args[0])
	return exitUsage
}

func namesAFile(args []string) ([]string, bool) {
	if len(args) != 1 {
		return nil, false
	}
	_, err := os.Stat(args[0])
	return args, err == nil
}

type sessionClock struct{}

func (sessionClock) Now() string { return time.Now().UTC().Format(time.RFC3339Nano) }
