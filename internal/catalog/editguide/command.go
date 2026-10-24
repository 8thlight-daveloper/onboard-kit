package editguide

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"

	contract "onboard-kit/contracts/catalog/editguide"
	editing "onboard-kit/contracts/catalog/editor"
	creation "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/catalog/editor"
	"onboard-kit/internal/catalog/materializer"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/authoringvalidator"
	"onboard-kit/internal/plan/decoder"
	"onboard-kit/internal/plan/encoder"
	"onboard-kit/internal/plan/validator"
)

type command struct {
	editing   editing.Editor
	creation  creation.Materializer
	decoding  decoder.Decoder
	encoding  encoder.Encoder
	authoring authoringvalidator.Validator
	checking  validator.Validator
}

func New() contract.Command {
	return command{editing: editor.New(), creation: materializer.New(), decoding: decoder.New(), encoding: encoder.New(), authoring: authoringvalidator.New(), checking: validator.New()}
}

const help = "Usage: onboard-kit catalog edit --plan FILE\n\nStage explicit ordered edits, review original and proposed plan bytes, then save.\nOnly the selected supported private plan may be replaced; personal files stay untouched.\nFind new source IDs with catalog list and catalog show --id ID.\n"
const startup = "Catalog edit draft — not saved.\nOnly this selected plan may be replaced after complete review and explicit save.\nNo record or preparation is read; saved personal answers and outcomes are not edited.\nExisting Submit empty values clear; initial Add bindings retain exact empty values.\nDo not enter secrets or unnecessary personal information. Scrollback may retain shared facts.\nInput limit: 1048576 bytes total; 65536 bytes per logical line.\nFind source IDs with catalog list and catalog show --id ID.\nPlan: "
const draftPrompt = "Edit draft — not saved. Choose add, submit, unset, retire, reactivate, drop, review, or cancel:\n"
const reviewPrompt = "Review complete. Choose save, back, or cancel:\n"
const invalidAction = "Choose one of the displayed actions.\n"

type terminal struct {
	input    io.Reader
	output   io.Writer
	consumed int
	failure  string
}

func write(output io.Writer, value string) bool {
	if output == nil {
		return false
	}
	n, err := io.WriteString(output, value)
	return err == nil && n == len(value)
}

func (screen *terminal) say(value string) bool {
	if screen.failure != "" {
		return false
	}
	if !write(screen.output, value) {
		screen.failure = "output"
		return false
	}
	return true
}

func (driver command) Run(args []string, input io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		if write(stdout, help) {
			return 0
		}
		return report(stderr, "output")
	}
	path, ok := arguments(args)
	if !ok {
		return report(stderr, "usage")
	}
	held, err := openSelected(path)
	if err != nil {
		return report(stderr, "path")
	}
	defer held.close()
	screen := &terminal{input: input, output: stdout}
	original, problems := driver.decoding.Decode(held.bytes)
	if problemErrors(problems) {
		if !screen.say(problemText(project(problems, "decode"))) {
			return report(stderr, "output")
		}
		return report(stderr, "data")
	}
	admitted, diagnostics := driver.editing.Edit(original, nil)
	if diagnosticErrors(diagnostics) {
		if !screen.say(problemText(diagnostics)) {
			return report(stderr, "output")
		}
		return report(stderr, "data")
	}
	_, diagnostics, faithful := driver.encodeFaithfully(admitted)
	if !faithful {
		if len(diagnostics) != 0 && !screen.say(problemText(diagnostics)) {
			return report(stderr, "output")
		}
		return report(stderr, "fidelity")
	}
	for _, stage := range []string{"authoring", "plan"} {
		var problems []model.Problem
		if stage == "authoring" {
			problems = driver.authoring.Validate(admitted)
		} else {
			problems = driver.checking.Validate(admitted)
		}
		if problemErrors(problems) {
			if !screen.say(problemText(project(problems, stage))) {
				return report(stderr, "output")
			}
			return report(stderr, "data")
		}
	}
	if !screen.say(startup + quote(held.path) + "\n") {
		return report(stderr, "output")
	}
	var operations []editing.Operation
	for screen.failure == "" {
		if !screen.say(draftPrompt) {
			break
		}
		action, complete := screen.line()
		if !complete {
			return screen.finish(stderr)
		}
		switch action {
		case "cancel":
			return report(stderr, "cancelled")
		case "drop":
			value, read := screen.field("Staged operation number:")
			if !read {
				return screen.finish(stderr)
			}
			number, valid := count(value, len(operations))
			if !valid || number == 0 {
				screen.say("Select an existing staged operation number.\n")
				continue
			}
			operations = append(operations[:number-1], operations[number:]...)
			screen.say("Operation dropped.\n")
		case "add", "submit", "unset", "retire", "reactivate":
			adds := 0
			for _, operation := range operations {
				if operation.Kind == editing.Add {
					adds++
				}
			}
			if len(operations) >= 1024 || action == "add" && adds >= 256 {
				screen.say("Edit limit reached; drop a staged operation before adding another.\n")
				continue
			}
			operation, read := screen.operation(action)
			if !read {
				return screen.finish(stderr)
			}
			operations = append(operations, operation)
			screen.say("Operation staged; validation occurs at review.\n")
		case "review":
			body, frozen, unchanged, diagnostics, available := driver.review(original, held.bytes, held.path, operations)
			if !available {
				if len(diagnostics) == 0 {
					screen.say("Review unavailable: preserve all plan data within supported limits.\n")
				} else {
					screen.say("Review unavailable.\n" + problemText(diagnostics))
				}
				continue
			}
			if !screen.say(body) {
				break
			}
			for screen.failure == "" {
				choice, read := screen.line()
				if !read {
					return screen.finish(stderr)
				}
				if choice == "cancel" {
					return report(stderr, "cancelled")
				}
				if choice == "back" {
					break
				}
				if choice == "save" {
					return report(stderr, held.commit(frozen, unchanged))
				}
				if !screen.say(invalidAction + reviewPrompt) {
					break
				}
			}
		default:
			screen.say(invalidAction)
		}
	}
	return screen.finish(stderr)
}

func arguments(args []string) (string, bool) {
	if len(args) < 1 || len(args) > 2 {
		return "", false
	}
	total := 0
	for _, arg := range args {
		total += len(arg)
		if !utf8.ValidString(arg) || strings.IndexByte(arg, 0) >= 0 {
			return "", false
		}
	}
	if total > 8192 {
		return "", false
	}
	path := ""
	if len(args) == 2 && args[0] == "--plan" && !strings.HasPrefix(args[1], "--") {
		path = args[1]
	}
	if len(args) == 1 && strings.HasPrefix(args[0], "--plan=") {
		path = strings.TrimPrefix(args[0], "--plan=")
	}
	return path, path != "" && len(path) <= 4096
}

func (driver command) encodeFaithfully(plan model.Plan) ([]byte, []editing.Diagnostic, bool) {
	encoded, problems := driver.encoding.Encode(plan)
	if problemErrors(problems) {
		return nil, project(problems, "encode"), false
	}
	if len(encoded) > 1048576 {
		return nil, nil, false
	}
	decoded, problems := driver.decoding.Decode(encoded)
	if problemErrors(problems) {
		return nil, project(problems, "decode"), false
	}
	if !equivalent(reflect.ValueOf(plan), reflect.ValueOf(decoded), false) {
		return nil, nil, false
	}
	return encoded, nil, true
}

func equivalent(left, right reflect.Value, unknown bool) bool {
	if !left.IsValid() || !right.IsValid() {
		return left.IsValid() == right.IsValid()
	}
	if left.Kind() == reflect.Interface || right.Kind() == reflect.Interface {
		if left.Kind() != reflect.Interface || right.Kind() != reflect.Interface {
			return false
		}
		if left.IsNil() || right.IsNil() {
			return left.IsNil() == right.IsNil()
		}
		return equivalent(left.Elem(), right.Elem(), true)
	}
	if left.Type() != right.Type() {
		return unknown && (left.Kind() == reflect.Int || left.Kind() == reflect.Int64) && (right.Kind() == reflect.Int || right.Kind() == reflect.Int64) && left.Int() == right.Int()
	}
	switch left.Kind() {
	case reflect.Pointer:
		if left.IsNil() || right.IsNil() {
			return left.IsNil() == right.IsNil()
		}
		return equivalent(left.Elem(), right.Elem(), unknown)
	case reflect.Map:
		if left.Len() != right.Len() || unknown && left.IsNil() != right.IsNil() {
			return false
		}
		for _, key := range left.MapKeys() {
			if !equivalent(left.MapIndex(key), right.MapIndex(key), unknown) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if left.Len() != right.Len() || unknown && left.IsNil() != right.IsNil() {
			return false
		}
		for i := 0; i < left.Len(); i++ {
			if !equivalent(left.Index(i), right.Index(i), unknown) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < left.NumField(); i++ {
			if !equivalent(left.Field(i), right.Field(i), unknown || left.Type().Field(i).Name == "UnknownFields") {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(left.Interface(), right.Interface())
	}
}

func problemErrors(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}
func diagnosticErrors(problems []editing.Diagnostic) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func project(problems []model.Problem, stage string) []editing.Diagnostic {
	var result []editing.Diagnostic
	for _, problem := range problems {
		message := map[string]string{
			"decode":    "Repair the selected plan's format; duplicate keys and invalid spellings must be fixed explicitly.",
			"encode":    "Preserve or explicitly repair unrepresentable plan data before saving.",
			"authoring": "Repair the selected plan's declarations before editing.",
			"plan":      "Repair the selected plan's structure or stored facts before editing.",
			"guidance":  "Ask the catalog maintainer to repair the advisory preparation source.",
		}[stage]
		switch problem.Code {
		case "directory.contact-unreachable":
			message = "Supply an approved business route for the declared contact."
		case "directory.role-vacant":
			message = "Name an accountable holder for the declared role."
		case "directory.email-malformed":
			message = "Repair the declared business email address."
		}
		if strings.HasPrefix(problem.Code, "secret.") {
			message = "Remove credential-like content before saving."
		}
		result = append(result, editing.Diagnostic{Stage: stage, Path: "plan", Code: problem.Code, Severity: problem.Severity, Message: message})
	}
	return result
}

func report(output io.Writer, code string) int {
	status, exit := "refused", 1
	switch code {
	case "cancelled", "unchanged", "written":
		status, exit = code, 0
	case "usage":
		exit = 2
	case "cleanup-written", "report-written":
		status = "written"
	}
	message := map[string]string{
		"cancelled":       "Cancelled. The selected plan was not changed.",
		"unchanged":       "The reviewed proposal leaves the selected plan unchanged.",
		"written":         "The reviewed plan edits were saved.",
		"usage":           "Use catalog edit --plan FILE.",
		"path":            "Select a safe single-link effective-owned0600 plan in an owned directory without group/other write; unsupported controls require separate review.",
		"data":            "Repair the displayed plan data before editing.",
		"fidelity":        "All held plan metadata must survive before this file can be edited.",
		"input":           "Use complete valid input within the displayed limits.",
		"output":          "Required review output failed; no plan replacement was accepted.",
		"busy":            "Another cooperating plan edit is saving; reload and review again.",
		"stale":           "The selected plan context changed; reload and review again.",
		"commit":          "The plan was not replaced; owned temporary cleanup may need repair.",
		"cleanup-refused": "The plan was not replaced; temporary cleanup needs repair.",
		"cleanup-written": "The plan edits were saved; temporary cleanup needs repair.",
		"report-written":  "The plan edits were saved, but reporting failed.",
	}[code]
	encoded, _ := json.Marshal(struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}{status, "edit." + code, message})
	if write(output, string(encoded)+"\n") {
		return exit
	}
	if status == "written" {
		fallback := "report-written"
		if code == "cleanup-written" {
			fallback = code
		}
		message = "The plan edits were saved, but reporting failed."
		if fallback == "cleanup-written" {
			message = "The plan edits were saved; temporary cleanup needs repair."
		}
		encoded, _ = json.Marshal(struct {
			Status  string `json:"status"`
			Code    string `json:"code"`
			Message string `json:"message"`
		}{"written", "edit." + fallback, message})
		write(output, string(encoded)+"\n")
	}
	return 1
}
