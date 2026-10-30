package editguide

import (
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"

	contract "onboard-kit/contracts/catalog/editguide"
	editorcontract "onboard-kit/contracts/catalog/editor"
	materializercontract "onboard-kit/contracts/catalog/materializer"
	authoringcontract "onboard-kit/contracts/plan/authoringvalidator"
	decodercontract "onboard-kit/contracts/plan/decoder"
	encodercontract "onboard-kit/contracts/plan/encoder"
	validatorcontract "onboard-kit/contracts/plan/validator"
	"onboard-kit/internal/catalog/editor"
	"onboard-kit/internal/catalog/materializer"
	"onboard-kit/internal/plan/authoringvalidator"
	"onboard-kit/internal/plan/decoder"
	"onboard-kit/internal/plan/encoder"
	"onboard-kit/internal/plan/validator"
)

const (
	maxArgumentBytes = 8192
	maxPathBytes     = 4096
	maxSourceBytes   = 1048576
)

const helpText = "Usage: onboard-kit catalog edit --plan FILE\n\n" +
	"Stage explicit ordered edits, review original and proposed plan bytes, then save.\n" +
	"Only the selected supported private plan may be replaced; personal files stay untouched.\n" +
	"Find new source IDs with catalog list and catalog show --id ID.\n"

type command struct {
	editor       editorcontract.Editor
	materializer materializercontract.Materializer
	decoder      decodercontract.Decoder
	encoder      encodercontract.Encoder
	authoring    authoringcontract.Validator
	checker      validatorcontract.Validator
}

func New() contract.Command {
	return command{
		editor:       editor.New(),
		materializer: materializer.New(),
		decoder:      decoder.New(),
		encoder:      encoder.New(),
		authoring:    authoringvalidator.New(),
		checker:      validator.New(),
	}
}

func (c command) Run(args []string, input io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		if writeAll(stdout, helpText) {
			return 0
		}
		return report(stderr, "output")
	}
	path, ok := planArgument(args)
	if !ok {
		return report(stderr, "usage")
	}
	chosen, code := selectPlan(path)
	if code != "" {
		return report(stderr, code)
	}
	original, problems, code := c.admit(chosen.raw)
	if code != "" {
		if problems != "" && !writeAll(stdout, problems) {
			code = "output"
		}
		return report(stderr, closeWith(chosen, code))
	}
	s := &session{command: c, input: input, output: stdout, selection: chosen, original: original}
	s.write(strings.ReplaceAll(startupText, "@PLAN@", quote(chosen.path)))
	code = s.run()
	return report(stderr, closeWith(chosen, code))
}

func closeWith(chosen *selection, code string) string {
	if chosen.close() == nil {
		return code
	}
	switch code {
	case "written", "cleanup-written":
		return "cleanup-written"
	case "cleanup-refused":
		return code
	}
	return "commit"
}

func planArgument(args []string) (string, bool) {
	if len(args) == 0 || len(args) > 2 {
		return "", false
	}
	total := 0
	for _, arg := range args {
		total += len(arg)
		if !utf8.ValidString(arg) {
			return "", false
		}
	}
	if total > maxArgumentBytes {
		return "", false
	}
	var path string
	if value, found := strings.CutPrefix(args[0], "--plan="); found {
		if len(args) != 1 {
			return "", false
		}
		path = value
	} else if args[0] == "--plan" && len(args) == 2 && !strings.HasPrefix(args[1], "--") {
		path = args[1]
	} else {
		return "", false
	}
	if path == "" || len(path) > maxPathBytes || strings.ContainsRune(path, 0) {
		return "", false
	}
	return path, true
}

func writeAll(writer io.Writer, text string) bool {
	if writer == nil {
		return false
	}
	n, err := io.WriteString(writer, text)
	return err == nil && n == len(text)
}

var reports = map[string]struct {
	status, message string
	exit            int
}{
	"cancelled":       {"cancelled", "Cancelled. The selected plan was not changed.", 0},
	"unchanged":       {"unchanged", "The reviewed proposal leaves the selected plan unchanged.", 0},
	"written":         {"written", "The reviewed plan edits were saved.", 0},
	"usage":           {"refused", "Use catalog edit --plan FILE.", 2},
	"path":            {"refused", "Select a safe single-link effective-owned0600 plan in an owned directory without group/other write; unsupported controls require separate review.", 1},
	"data":            {"refused", "Repair the displayed plan data before editing.", 1},
	"fidelity":        {"refused", "All held plan metadata must survive before this file can be edited.", 1},
	"input":           {"refused", "Use complete valid input within the displayed limits.", 1},
	"output":          {"refused", "Required review output failed; no plan replacement was accepted.", 1},
	"busy":            {"refused", "Another cooperating plan edit is saving; reload and review again.", 1},
	"stale":           {"refused", "The selected plan context changed; reload and review again.", 1},
	"commit":          {"refused", "The plan was not replaced; owned temporary cleanup may need repair.", 1},
	"cleanup-refused": {"refused", "The plan was not replaced; temporary cleanup needs repair.", 1},
	"cleanup-written": {"written", "The plan edits were saved; temporary cleanup needs repair.", 1},
	"report-written":  {"written", "The plan edits were saved, but reporting failed.", 1},
}

func report(writer io.Writer, code string) int { return reportAttempt(writer, code, true) }

func reportAttempt(writer io.Writer, code string, retry bool) int {
	row := reports[code]
	encoded, _ := json.Marshal(struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}{row.status, "edit." + code, row.message})
	if writeAll(writer, string(encoded)+"\n") {
		return row.exit
	}
	if row.status == "written" && retry {
		if code == "cleanup-written" {
			reportAttempt(writer, "cleanup-written", false)
		} else {
			reportAttempt(writer, "report-written", false)
		}
	}
	return 1
}
