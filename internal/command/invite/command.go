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
	request, help, valid := parse(args)
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
	result := Export(request)
	report := result.diagnostics()
	if !result.Written() {
		return refuse(stderr, report)
	}
	stdoutOK := writeText(stdout, "Invitation archive written.\n")
	if !stdoutOK && result.Code == codeWritten {
		report.code = "result-output"
	}
	stderrOK := writeReport(stderr, report)
	if !stdoutOK || !stderrOK || result.Code != codeWritten {
		return 1
	}
	return 0
}

func refuse(stderr io.Writer, report diagnostics) int {
	writeReport(stderr, report)
	return 1
}
