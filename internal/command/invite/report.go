package invite

import (
	"encoding/json"
	"io"

	"onboard-kit/internal/model"
)

type diagnostics struct {
	code     string
	problems []model.Problem
	blockers []model.Blocker
}

type privateProblem struct {
	Code     string         `json:"code"`
	Severity model.Severity `json:"severity"`
}

type privateReport struct {
	Status   string           `json:"status"`
	Code     string           `json:"code"`
	Message  string           `json:"message"`
	Problems []privateProblem `json:"problems"`
	Blockers []string         `json:"blockers"`
}

func writeText(writer io.Writer, text string) bool {
	n, err := io.WriteString(writer, text)
	return err == nil && n == len(text)
}

func writeReport(writer io.Writer, report diagnostics) bool {
	value := privateReport{
		Status: "refused", Code: "invite." + report.code, Message: reportMessage(report.code),
		Problems: make([]privateProblem, 0, len(report.problems)), Blockers: make([]string, 0, len(report.blockers)),
	}
	if report.code == "written" || report.code == "result-output" || report.code == "cleanup-written" {
		value.Status = "written"
	}
	for _, problem := range report.problems {
		value.Problems = append(value.Problems, privateProblem{Code: problem.Code, Severity: problem.Severity})
	}
	for _, blocker := range report.blockers {
		value.Blockers = append(value.Blockers, blocker.Question)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return false
	}
	return writeText(writer, string(encoded)+"\n")
}

func reportMessage(code string) string {
	switch code {
	case "usage":
		return "Use invite export with all required explicit flags."
	case "account":
		return "Supply a valid explicit invitation account within the path budget."
	case "output-path":
		return "Choose a new archive in an existing nonsymlink directory."
	case "output-exists":
		return "Choose a new archive pathname; existing output is never replaced."
	case "plan-read":
		return "Supply a bounded local nonsymlink regular plan file."
	case "preparation-read":
		return "Supply a bounded local nonsymlink regular preparation file."
	case "plan-decode":
		return "Correct the plan file's format before export."
	case "preparation-decode":
		return "Correct the preparation file's format before export."
	case "plan-identity":
		return "Give the organization plan a nonempty identity before export."
	case "plan-invalid":
		return "Inspect the plan with the supported stdio validate workflow and correct its validation errors before export."
	case "preparation-invalid":
		return "Correct the preparation identity or details before export."
	case "onboarder-blocked":
		return "Supply or deliberately defer the required onboarder answers before export."
	case "owner":
		return "Name the onboarding owner in preparation with a usable display name."
	case "preaccess-help":
		return "Designate a known named business contact with a displayable pre-access help route."
	case "firststeps":
		return "Correct the help selection or first-steps inputs before export."
	case "plan-encode":
		return "Correct plan values that cannot be encoded before export."
	case "preparation-encode":
		return "Correct preparation values that cannot be encoded before export."
	case "record-encode":
		return "The identity-only record seed could not be encoded."
	case "archive":
		return "Keep all invitation payloads within the archive limits."
	case "publish":
		return "The new archive could not be safely published; existing files were preserved."
	case "cleanup-refused":
		return "No archive was published, but temporary invitation data may remain. Restore access to the selected output directory and check exporter-owned temporary data before retrying."
	case "cleanup-written":
		return "The archive was written, but temporary invitation data may remain. Keep the archive; restore access to the selected output directory and check exporter-owned temporary data."
	case "written":
		return "Pre-access help availability has not been tested; deliver through an approved channel."
	case "result-output":
		return "The archive was written, but its result could not be fully displayed."
	case "help-output":
		return "The usage text could not be displayed."
	default:
		return "Use invite export with all required explicit flags."
	}
}
