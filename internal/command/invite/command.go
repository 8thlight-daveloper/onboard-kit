package invite

import (
	"io"

	contract "onboard-kit/contracts/command/invite"
)

type command struct{}

func New() contract.Command {
	return command{}
}

func (command) Run(args []string, stdout, stderr io.Writer) int {
	options, help, valid := parse(args)
	if help {
		if !writeText(stdout, usage) {
			writeReport(stderr, diagnostics{code: "help-output"})
			return 1
		}
		return 0
	}
	if !valid {
		if !writeReport(stderr, diagnostics{code: "usage"}) {
			return 1
		}
		return 2
	}
	if !validAccount(options.account) {
		return refuse(stderr, diagnostics{code: "account"})
	}
	output, code := openDestination(options.output)
	if code != "" {
		return refuse(stderr, diagnostics{code: code})
	}
	defer output.close()
	plan, ok := readSource(options.plan)
	if !ok {
		return refuse(stderr, diagnostics{code: "plan-read"})
	}
	preparation, ok := readSource(options.preparation)
	if !ok {
		return refuse(stderr, diagnostics{code: "preparation-read"})
	}
	archive, report := buildArchive(options, plan, preparation)
	if report.code != "" {
		return refuse(stderr, report)
	}
	publication := output.publish(archive)
	if !publication.committed {
		report.code = publication.code
		return refuse(stderr, report)
	}
	report.code = "written"
	if publication.code != "" {
		report.code = publication.code
	}
	stdoutOK := writeText(stdout, "Invitation archive written.\n")
	if !stdoutOK && publication.code == "" {
		report.code = "result-output"
	}
	stderrOK := writeReport(stderr, report)
	if !stdoutOK || !stderrOK || publication.code != "" {
		return 1
	}
	return 0
}

func refuse(stderr io.Writer, report diagnostics) int {
	writeReport(stderr, report)
	return 1
}
