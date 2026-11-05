package create

import (
	"flag"
	"fmt"
	"io"
	"os"

	"onboard-kit/internal/command/planfile"
)

const (
	exitSuccess = 0
	exitProblem = 1
	exitUsage   = 2
)

type Options struct {
	Templates        string
	Output           string
	TemplateSet      string
	TemplateSetGiven bool
	GeneratedOn      string
	Identifier       string
}

func Parse(args []string, stderr io.Writer) (Options, bool) {
	options := flag.NewFlagSet("new", flag.ContinueOnError)
	options.SetOutput(stderr)
	templates := options.String("templates", "", "folder of section template files, and an optional plan-variables.yaml, kept for existing scripts")
	output := options.String("out", "", "file to write, standard output when left out and no terminal is present")
	templateSet := options.String("template-set", "local", "name recorded as the template set, with --templates only")
	generatedOn := options.String("generated-on", "", "date recorded in the plan, today when left out")
	identifier := options.String("id", "", "identity of the plan, a new random one when left out")
	if options.Parse(args) != nil {
		return Options{}, false
	}
	if options.NArg() > 0 {
		fmt.Fprintf(stderr, "new: unexpected argument %q\n", options.Arg(0))
		return Options{}, false
	}
	parsed := Options{Templates: *templates, Output: *output, TemplateSet: *templateSet, GeneratedOn: *generatedOn, Identifier: *identifier}
	options.Visit(func(given *flag.Flag) {
		if given.Name == "template-set" {
			parsed.TemplateSetGiven = true
		}
	})
	return parsed, true
}

func FromTemplates(options Options, stdout io.Writer, stderr io.Writer) int {
	plan, ok := planfile.FromTemplates(options.Templates, options.TemplateSet, options.GeneratedOn, stderr)
	if !ok {
		return exitProblem
	}
	plan.ID = options.Identifier
	if plan.ID == "" {
		plan.ID = NewIdentifier()
	}
	data, ok := planfile.Encode(plan, stderr)
	if !ok {
		return exitProblem
	}
	return write(options.Output, data, stdout, stderr)
}

func write(path string, data []byte, stdout io.Writer, stderr io.Writer) int {
	if path == "" {
		if _, err := stdout.Write(data); err != nil {
			fmt.Fprintln(stderr, err)
			return exitProblem
		}
		return exitSuccess
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return exitProblem
	}
	return exitSuccess
}
