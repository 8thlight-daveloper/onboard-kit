package frontdoor

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/x/term"

	starterapi "onboard-kit/contracts/invitation/starter"
	"onboard-kit/internal/command/create"
	"onboard-kit/internal/command/report"
	"onboard-kit/internal/command/walk"
	"onboard-kit/internal/invitation/starter"
	"onboard-kit/internal/tui"
)

const (
	exitSuccess     = 0
	exitProblem     = 1
	exitUsage       = 2
	exitInterrupted = 130
	exitTerminated  = 143

	ScriptingCommands  = "For scripts, use new --out FILE, assemble, catalog, prepare, invite, or stdio."
	editorNeedsTerm    = "onboard-kit: the plan editor needs a terminal. For scripts, use assemble --plan FILE or stdio."
	onboardingNeedTerm = "onboard-kit: beginning an onboarding needs a terminal. For scripts, use start --invitation FILE --root DIR."
	recordsNotice      = "Your onboarding records are kept in %s, the folder you started onboard-kit in.\n"
	typedNotice        = "The screens for your onboarding are not built yet, so it goes on with typed commands.\n"
)

func Terminal(stdin io.Reader, stdout io.Writer) bool {
	input, inputIsFile := stdin.(*os.File)
	output, outputIsFile := stdout.(*os.File)
	return inputIsFile && outputIsFile && term.IsTerminal(input.Fd()) && term.IsTerminal(output.Fd())
}

func Run(version string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	door := newPort(version, "", "")
	choice, err := tui.RunFrontDoor(tui.NewFrontDoor(door), stdin, stdout)
	return act(choice, err, door, version, stdin, stdout, stderr)
}

func Open(args []string, version string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: onboard-kit open FILE")
		return exitUsage
	}
	checked, problems := Check(args[0])
	if len(problems) > 0 {
		report.Problems(stderr, args[0], problems)
		return exitProblem
	}
	if checked.Kind == tui.FileRefused {
		fmt.Fprintln(stderr, "onboard-kit: "+checked.Refusal)
		return exitProblem
	}
	terminal := Terminal(stdin, stdout)
	switch {
	case checked.Kind == tui.FilePlan && !terminal:
		fmt.Fprintln(stderr, editorNeedsTerm)
		return exitUsage
	case checked.Kind == tui.FilePlan:
		return walk.Open(args[0], version, stdin, stdout, stderr)
	case !terminal:
		fmt.Fprintln(stderr, onboardingNeedTerm)
		return exitUsage
	}
	return begin(checked.Path, newPort(version, "", "").working, stdin, stdout, stderr)
}

func New(args []string, version string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	options, ok := create.Parse(args, stderr)
	if !ok {
		return exitUsage
	}
	if options.Templates != "" {
		return create.FromTemplates(options, stdout, stderr)
	}
	if options.TemplateSetGiven {
		fmt.Fprintln(stderr, "new: --template-set applies only with --templates. A new plan records the version of this tool.")
		return exitUsage
	}
	if Terminal(stdin, stdout) {
		door := newPort(version, options.Identifier, options.GeneratedOn)
		path := options.Output
		if path == "" {
			path = door.ProposedPlanPath()
		}
		choice, err := tui.RunFrontDoor(tui.NewPlanDoor(door, path), stdin, stdout)
		return act(choice, err, door, version, stdin, stdout, stderr)
	}
	return newForScripts(options, version, stdout, stderr)
}

func newForScripts(options create.Options, version string, stdout io.Writer, stderr io.Writer) int {
	data, refusal := corePlanBytes(options.Identifier, version, options.GeneratedOn)
	if refusal != "" {
		fmt.Fprintln(stderr, "new: "+refusal)
		return exitProblem
	}
	if options.Output == "" {
		if _, err := stdout.Write(data); err != nil {
			fmt.Fprintln(stderr, err)
			return exitProblem
		}
		return exitSuccess
	}
	err := create.WriteNew(options.Output, data)
	switch {
	case errors.Is(err, create.ErrExists):
		fmt.Fprintf(stderr, "new: %s already exists, and a new plan never replaces a file.\n", options.Output)
		return exitProblem
	case err != nil:
		fmt.Fprintln(stderr, "new:", err)
		return exitProblem
	}
	return exitSuccess
}

func act(choice tui.FrontDoorChoice, err error, door *port, version string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	switch {
	case errors.Is(err, tui.ErrInterrupted):
		return exitInterrupted
	case errors.Is(err, tui.ErrTerminated):
		return exitTerminated
	case err != nil:
		fmt.Fprintln(stderr, err)
		return exitProblem
	}
	switch choice.Action {
	case tui.FrontDoorOpenPlan:
		return walk.Open(choice.Path, version, stdin, stdout, stderr)
	case tui.FrontDoorBeginOnboarding:
		return begin(choice.Path, door.working, stdin, stdout, stderr)
	}
	return exitSuccess
}

func begin(invitation string, root string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	fmt.Fprintf(stdout, recordsNotice, root)
	fmt.Fprint(stdout, typedNotice)
	return starter.New(starterapi.Runtime{}).Run([]string{"--invitation", invitation, "--root", root}, stdin, stdout, stderr)
}
