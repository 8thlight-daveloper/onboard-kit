package root

import (
	"io"
	"strings"

	starterapi "onboard-kit/contracts/invitation/starter"
	legacy "onboard-kit/internal/command/start"
	"onboard-kit/internal/invitation/starter"
)

func startCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return starter.New(starterapi.Runtime{}).Run(args, stdin, stdout, stderr)
	}
	invitationMode, legacyMode := false, false
	for index := 0; index < len(args); index++ {
		if args[index] == "--" {
			break
		}
		key, _, equals := strings.Cut(args[index], "=")
		switch key {
		case "--invitation", "--resume-import":
			invitationMode = true
		case "--org-plan", "-org-plan", "--no-walk", "-no-walk", "--refreshed-on", "-refreshed-on":
			legacyMode = true
		}
		if !equals {
			switch key {
			case "--invitation", "--resume-import", "--root", "--home", "-home", "--org-plan", "-org-plan", "--refreshed-on", "-refreshed-on":
				index++
			}
		}
	}
	if legacyMode && !invitationMode {
		return legacy.Run(args, stdin, stdout, stderr)
	}
	command := starter.New(starterapi.Runtime{})
	if legacyMode {
		return command.Run(nil, stdin, stdout, stderr)
	}
	return command.Run(args, stdin, stdout, stderr)
}
