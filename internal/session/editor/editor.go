package editor

import (
	"io"
	"strings"
	"unicode/utf8"

	contract "onboard-kit/contracts/session/editor"
	reviewcontract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/decoder"
	"onboard-kit/internal/preparation/locator"
	"onboard-kit/internal/session/reviewer"
)

type Driver struct{ clock contract.Clock }

func New(clock contract.Clock) Driver { return Driver{clock: clock} }

const help = "Usage: onboard-kit session edit --root ABSROOT --home ABSHOME --session KEY\n\nReview one explicit newcomer answer, follow-plan choice, or outcome and separately save the selected record.\nNo commands or checks run. Plan, preparation, registration and first-use data are retained.\nInterrupted private staging is preserved and requires repair; this command never replays or automatically deletes it.\nNative preservation may refuse unsupported metadata. Atomic record replacement is not power-loss or hostile-owner protection.\n"

func report(w io.Writer, code string, saved bool) contract.Result {
	status, message, exit := "refused", "", 1
	switch code {
	case "usage":
		exit = 2
		message = "Use an explicit root, state home and registered session key."
	case "cancelled":
		exit = 0
		status = "cancelled"
		message = "Session change cancelled; no record was written."
	case "unchanged":
		exit = 0
		status = "unchanged"
		message = "The record is unchanged."
	case "saved":
		exit = 0
		status = "saved"
		message = "The selected record change was saved."
	case "context":
		message = "Select a safe unchanged registered session context."
	case "pending":
		message = "Finish or repair the pending import first."
	case "source":
		message = "Read the bounded existing session files before editing."
	case "review":
		message = "Resolve the displayed session problem before saving."
	case "input":
		message = "Use complete valid displayed actions within the input limits."
	case "output":
		message = "Required output could not be completed; no record was written."
	case "limit":
		message = "The selected session or display exceeds the supported limit."
	case "commit":
		message = "The record was not replaced; private staging may need repair."
	case "repair":
		status = "saved"
		message = "The record was saved, but cleanup or reporting needs repair."
	}
	flag := "false"
	if saved {
		flag = "true"
	}
	line := "{\"status\":\"" + status + "\",\"code\":\"edit." + code + "\",\"message\":\"" + message + "\",\"saved\":" + flag + "}\n"
	if !writeBlock(w, line) {
		exit = 1
		if saved && code != "repair" {
			_ = report(w, "repair", true)
		}
	}
	return contract.Result{Exit: exit, Saved: saved}
}

func (d Driver) Run(arguments []string, streams contract.Streams) contract.Result {
	if len(arguments) == 1 && arguments[0] == "--help" {
		if !writeBlock(streams.Output, help) {
			return report(streams.Error, "output", false)
		}
		return contract.Result{}
	}
	args, ok := parseArguments(arguments)
	if !ok {
		return report(streams.Error, "usage", false)
	}
	if d.clock == nil {
		return report(streams.Error, "context", false)
	}
	state, code := openContext(args)
	if code != "" {
		return report(streams.Error, code, false)
	}
	defer state.close()
	input := lineReader{reader: streams.Input}
	block := identity(state)
	if block.failed {
		return report(streams.Error, "limit", false)
	}
	if !writeBlock(streams.Output, block.String()) {
		return report(streams.Error, "output", false)
	}
	line, complete, valid := input.line()
	if !valid {
		return report(streams.Error, "input", false)
	}
	if !complete || line == "cancel" {
		return report(streams.Error, "cancelled", false)
	}
	if line != "confirm "+quote(state.selected.account) {
		return report(streams.Error, "input", false)
	}
	source, code := state.load()
	if code != "" {
		return report(streams.Error, code, false)
	}
	core := reviewer.New()
	session, problems := core.Inspect(source)
	if fatal(problems) {
		block = builder{}
		block.diagnostics("Session problems", problems)
		if block.failed {
			return report(streams.Error, "limit", false)
		}
		if !writeBlock(streams.Output, block.String()) {
			return report(streams.Error, "output", false)
		}
		return report(streams.Error, reviewCode(problems), false)
	}
	block = displayCurrent(session, problems)
	if block.failed {
		return report(streams.Error, "limit", false)
	}
	if !writeBlock(streams.Output, block.String()) {
		return report(streams.Error, "output", false)
	}
	line, complete, valid = input.line()
	if !valid {
		return report(streams.Error, "input", false)
	}
	if !complete || line == "exit" {
		return report(streams.Error, "cancelled", false)
	}
	intent, ok := parseIntent(line)
	if !ok {
		return report(streams.Error, "input", false)
	}
	proposal, problems := core.Propose(source, intent)
	if fatal(problems) {
		block = builder{}
		block.diagnostics("Session problems", problems)
		if block.failed {
			return report(streams.Error, "limit", false)
		}
		if !writeBlock(streams.Output, block.String()) {
			return report(streams.Error, "output", false)
		}
		return report(streams.Error, reviewCode(problems), false)
	}
	block = displayProposal(proposal, problems)
	if block.failed {
		return report(streams.Error, "limit", false)
	}
	if !writeBlock(streams.Output, block.String()) {
		return report(streams.Error, "output", false)
	}
	line, complete, valid = input.line()
	if !valid {
		return report(streams.Error, "input", false)
	}
	if !complete || line == "cancel" {
		return report(streams.Error, "cancelled", false)
	}
	if line != "save" {
		return report(streams.Error, "input", false)
	}
	if code = state.fresh(); code != "" {
		return report(streams.Error, code, false)
	}
	event := model.Timestamp("")
	if proposal.NeedsEvent {
		event = model.Timestamp(d.clock.Now())
	}
	publication, problems := core.Finalize(proposal, event)
	if fatal(problems) {
		block = builder{}
		block.diagnostics("Session problems", problems)
		if block.failed {
			return report(streams.Error, "limit", false)
		}
		if !writeBlock(streams.Output, block.String()) {
			return report(streams.Error, "output", false)
		}
		return report(streams.Error, reviewCode(problems), false)
	}
	if !publication.Changed {
		if !state.unlock() {
			return report(streams.Error, "context", false)
		}
		if !writeBlock(streams.Output, "Record unchanged.\n") {
			return report(streams.Error, "output", false)
		}
		return report(streams.Error, "unchanged", false)
	}
	saved, clean := state.publish(publication.Record)
	if !saved {
		return report(streams.Error, "commit", false)
	}
	if !state.close() {
		clean = false
	}
	if !clean {
		_ = writeBlock(streams.Output, "Record saved.\n")
		return report(streams.Error, "repair", true)
	}
	if !writeBlock(streams.Output, "Record saved.\n") {
		return report(streams.Error, "repair", true)
	}
	return report(streams.Error, "saved", true)
}

func fatal(problems []reviewcontract.Diagnostic) bool {
	for _, p := range problems {
		if p.Problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}
func reviewCode(problems []reviewcontract.Diagnostic) string {
	for _, p := range problems {
		if p.Problem.Code == "session.resource" {
			return "limit"
		}
	}
	return "review"
}

type arguments struct{ root, home, key string }

func parseArguments(args []string) (arguments, bool) {
	result := arguments{}
	seen := map[string]bool{}
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) || seen[args[i]] || args[i+1] == "" || strings.HasPrefix(args[i+1], "--") {
			return arguments{}, false
		}
		seen[args[i]] = true
		switch args[i] {
		case "--root":
			result.root = args[i+1]
		case "--home":
			result.home = args[i+1]
		case "--session":
			result.key = args[i+1]
		default:
			return arguments{}, false
		}
	}
	return result, len(seen) == 3 && absolute(result.root) && absolute(result.home) && hexKey(result.key)
}
func absolute(path string) bool {
	if len(path) == 0 || len(path) > 4096 || path[0] != '/' || !utf8.ValidString(path) || strings.ContainsRune(path, 0) {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return false
		}
	}
	return true
}
func hexKey(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func boundSource(plan []byte, account, planID string, preparation, record []byte) reviewcontract.Source {
	return reviewcontract.Source{Plan: plan, Account: account, PlanID: planID, Preparation: preparation, Record: record}
}
func locate(plan []byte, account string) (string, string, bool) {
	p, problems := decoder.New().Decode(plan)
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return "", "", false
		}
	}
	paths, problems := locator.New().Locate(p.ResultsPath, account)
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return "", "", false
		}
	}
	return paths.Record, paths.Preparation, true
}
