package creator

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	contract "onboard-kit/contracts/catalog/creator"
	materializercontract "onboard-kit/contracts/catalog/materializer"
	encodercontract "onboard-kit/contracts/plan/encoder"
	"onboard-kit/internal/catalog/materializer"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/encoder"
)

const draftPrompt = "Draft — not saved. Choose add, bind, unbind, remove, review, or cancel:\n"
const reviewPrompt = "Review complete. Choose save, back, or cancel:\n"
const invalidAction = "Choose one of the displayed actions.\n"
const helpText = "Usage: onboard-kit catalog create --id ID --out FILE\n\n" +
	"Stage explicit activity instances and shared facts, review, and save one new plan.\n" +
	"Existing plans, preparation and newcomer records are not changed.\n" +
	"Choose sources with catalog list and catalog show --id ID.\n"
const startupText = "Catalog draft — not saved.\n" +
	"Shared organization facts only; no personal preparation or records are read or changed.\n" +
	"Do not enter secrets or unnecessary personal information. Terminal scrollback can retain entered facts.\n" +
	"Interaction limit: 1048576 input bytes total; 65536 bytes per logical line. No partial save.\n" +
	"Choose source IDs with catalog list; inspect content with catalog show --id ID.\n"

type command struct {
	materializer materializercontract.Materializer
	encoder      encodercontract.Encoder
}

func New() contract.Command {
	return command{materializer: materializer.New(), encoder: encoder.New()}
}

type session struct {
	command  command
	input    io.Reader
	output   io.Writer
	consumed int
	failure  string
	ended    bool
	request  materializercontract.Request
	result   materializercontract.Result
	parent   *destination
}

func (c command) Run(args []string, input io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		if writeAll(stdout, helpText) {
			return 0
		}
		return report(stderr, "output")
	}
	id, path, ok := arguments(args)
	if !ok {
		return report(stderr, "usage")
	}
	parent, err := admitDestination(path)
	if err != nil {
		return report(stderr, "path")
	}
	s := session{command: c, input: input, output: stdout, parent: parent,
		request: materializercontract.Request{PlanID: id, Instances: make([]materializercontract.Instance, 0)}}
	result, problems := c.materializer.Create(s.request)
	if hasErrors(problems) {
		if !s.write(problemText(problems)) {
			parent.close()
			return report(stderr, "output")
		}
		parent.close()
		return report(stderr, "data")
	}
	s.result = result
	s.write(startupText + "Plan: " + quote(id) + "\nDestination: " + quote(parent.path) + "\n" + draftPrompt)
	code := s.run()
	if parent.close() != nil {
		if code == "written" || code == "cleanup-written" {
			code = "cleanup-written"
		} else if code != "cleanup-refused" {
			code = "commit"
		}
	}
	return report(stderr, code)
}

func arguments(args []string) (string, string, bool) {
	if len(args) > 4 {
		return "", "", false
	}
	held := 0
	for _, arg := range args {
		if len(arg) > 32768-held {
			return "", "", false
		}
		held += len(arg)
	}
	values := map[string]string{}
	for i := 0; i < len(args); i++ {
		key, value, equals := strings.Cut(args[i], "=")
		if key != "--id" && key != "--out" {
			return "", "", false
		}
		if _, exists := values[key]; exists {
			return "", "", false
		}
		if !equals {
			i++
			if i == len(args) || strings.HasPrefix(args[i], "--") {
				return "", "", false
			}
			value = args[i]
		}
		if value == "" || len(value) > 4096 || !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
			return "", "", false
		}
		values[key] = value
	}
	id, idOK := values["--id"]
	path, pathOK := values["--out"]
	return id, path, idOK && pathOK
}

func (s *session) write(text string) bool {
	if s.failure != "" {
		return false
	}
	if !writeAll(s.output, text) {
		s.failure = "output"
		return false
	}
	return true
}

func writeAll(writer io.Writer, text string) bool {
	if writer == nil {
		return false
	}
	n, err := io.WriteString(writer, text)
	return err == nil && n == len(text)
}

func (s *session) line() (string, bool) {
	if s.failure != "" || s.ended {
		return "", false
	}
	data := make([]byte, 0, 128)
	for {
		if s.consumed == 1048576 {
			s.failure = "input"
			return "", false
		}
		var one [1]byte
		n, err := s.input.Read(one[:])
		if n < 0 || n > 1 {
			s.failure = "input"
			return "", false
		}
		s.consumed += n
		if err != nil && err != io.EOF {
			s.failure = "input"
			return "", false
		}
		if n == 0 {
			if err == io.EOF {
				if !utf8.Valid(data) {
					s.failure = "input"
				} else {
					s.ended = true
				}
				return "", false
			}
			s.failure = "input"
			return "", false
		}
		if one[0] == '\n' {
			if len(data) > 0 && data[len(data)-1] == '\r' {
				data = data[:len(data)-1]
			}
			if len(data) > 65536 || !utf8.Valid(data) {
				s.failure = "input"
				return "", false
			}
			return string(data), true
		}
		data = append(data, one[0])
		if len(data) > 65537 || (len(data) == 65537 && one[0] != '\r') {
			s.failure = "input"
			return "", false
		}
		if err == io.EOF {
			if len(data) > 65536 || !utf8.Valid(data) {
				s.failure = "input"
			} else {
				s.ended = true
			}
			return "", false
		}
	}
}

func (s *session) ask(prompt string) (string, bool) {
	if !s.write(prompt + "\n") {
		return "", false
	}
	return s.line()
}

func (s *session) run() string {
	for {
		action, ok := s.line()
		if !ok {
			if s.failure != "" {
				return s.failure
			}
			return "cancelled"
		}
		switch action {
		case "cancel":
			return "cancelled"
		case "add":
			s.add()
		case "bind":
			s.bind()
		case "unbind":
			s.unbind()
		case "remove":
			s.remove()
		case "review":
			if bytes, ok := s.review(); ok {
				for {
					choice, read := s.line()
					if !read {
						if s.failure != "" {
							return s.failure
						}
						return "cancelled"
					}
					if choice == "cancel" {
						return "cancelled"
					}
					if choice == "save" {
						return s.parent.publish(bytes)
					}
					if choice == "back" {
						break
					}
					if !s.write(invalidAction + reviewPrompt) {
						return s.failure
					}
				}
			}
		default:
			s.write(invalidAction)
		}
		if !s.write(draftPrompt) {
			return s.failure
		}
	}
}

func cloneRequest(request materializercontract.Request) materializercontract.Request {
	result := request
	result.Instances = make([]materializercontract.Instance, len(request.Instances))
	for i, instance := range request.Instances {
		result.Instances[i] = instance
		result.Instances[i].Bindings = append([]materializercontract.Binding(nil), instance.Bindings...)
	}
	return result
}

func (s *session) apply(candidate materializercontract.Request) {
	result, problems := s.command.materializer.Create(candidate)
	if hasErrors(problems) {
		s.write("Change not applied.\n" + problemText(problems))
		return
	}
	s.request = candidate
	s.result = result
	s.write("Change staged.\n")
}

func (s *session) add() {
	source, ok := s.ask("Source ID:")
	if !ok {
		return
	}
	id, ok := s.ask("Permanent section ID:")
	if !ok {
		return
	}
	candidate := cloneRequest(s.request)
	candidate.Instances = append(candidate.Instances, materializercontract.Instance{SourceID: source, PermanentSectionID: id})
	s.apply(candidate)
}

func (s *session) instance() (int, bool) {
	id, ok := s.ask("Instance ID:")
	if !ok {
		return 0, false
	}
	for i, instance := range s.request.Instances {
		if instance.PermanentSectionID == id {
			return i, true
		}
	}
	s.write("Select an existing draft instance.\n")
	return 0, false
}

func (s *session) bind() {
	i, ok := s.instance()
	if !ok {
		return
	}
	id := s.request.Instances[i].PermanentSectionID
	for _, section := range s.result.Plan.Sections {
		if section.ID != id {
			continue
		}
		var display strings.Builder
		fmt.Fprintf(&display, "Variables: %d\n", len(section.Variables))
		for _, variable := range section.Variables {
			kind, audience := variable.Type, variable.Audience
			if kind == "" {
				kind = model.DefaultVariableType
			}
			if audience == "" {
				audience = model.DefaultAudience
			}
			fmt.Fprintf(&display, "  - Name: %s Type: %s Audience: %s Required: %t\n", quote(variable.Name), quote(string(kind)), quote(string(audience)), variable.Required)
		}
		if !s.write(display.String()) {
			return
		}
		break
	}
	name, ok := s.ask("Variable name:")
	if !ok {
		return
	}
	value, ok := s.value()
	if !ok {
		return
	}
	candidate := cloneRequest(s.request)
	bindings := candidate.Instances[i].Bindings
	replaced := false
	for j := range bindings {
		if bindings[j].Name == name {
			bindings[j].Value = value
			replaced = true
			break
		}
	}
	if !replaced {
		bindings = append(bindings, materializercontract.Binding{Name: name, Value: value})
	}
	candidate.Instances[i].Bindings = bindings
	s.apply(candidate)
}

func (s *session) value() (model.Value, bool) {
	for {
		kind, ok := s.ask("Value kind: text, flag, list, or deferred:")
		if !ok {
			return model.Value{}, false
		}
		switch kind {
		case "text":
			text, ok := s.ask("Text value (empty text is explicit):")
			return model.Value{Kind: model.ValueKindText, Text: text}, ok
		case "deferred":
			return model.Value{Kind: model.ValueKindDeferred}, true
		case "flag":
			for {
				flag, ok := s.ask("Flag value: true or false:")
				if !ok {
					return model.Value{}, false
				}
				if flag == "true" || flag == "false" {
					return model.Value{Kind: model.ValueKindFlag, Flag: flag == "true"}, true
				}
				if !s.write("Use true or false.\n") {
					return model.Value{}, false
				}
			}
		case "list":
			count := 0
			for {
				text, ok := s.ask("List length: 0-4096:")
				if !ok {
					return model.Value{}, false
				}
				n, err := strconv.Atoi(text)
				if err == nil && n >= 0 && n <= 4096 && strconv.Itoa(n) == text {
					count = n
					break
				}
				if !s.write("Use a canonical count from 0 through 4096.\n") {
					return model.Value{}, false
				}
			}
			items := make([]string, 0, count)
			for j := 0; j < count; j++ {
				text, ok := s.ask(fmt.Sprintf("List item %d:", j+1))
				if !ok {
					return model.Value{}, false
				}
				items = append(items, text)
			}
			return model.Value{Kind: model.ValueKindList, List: items}, true
		default:
			if !s.write(invalidAction) {
				return model.Value{}, false
			}
		}
	}
}

func (s *session) unbind() {
	i, ok := s.instance()
	if !ok {
		return
	}
	name, ok := s.ask("Variable name:")
	if !ok {
		return
	}
	for j, binding := range s.request.Instances[i].Bindings {
		if binding.Name != name {
			continue
		}
		candidate := cloneRequest(s.request)
		bindings := candidate.Instances[i].Bindings
		candidate.Instances[i].Bindings = append(bindings[:j], bindings[j+1:]...)
		s.apply(candidate)
		return
	}
	s.write("Select an existing draft binding.\n")
}

func (s *session) remove() {
	i, ok := s.instance()
	if !ok {
		return
	}
	candidate := cloneRequest(s.request)
	candidate.Instances = append(candidate.Instances[:i], candidate.Instances[i+1:]...)
	s.apply(candidate)
}

func hasErrors(problems []materializercontract.Diagnostic) bool {
	for _, problem := range problems {
		if problem.Problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func report(writer io.Writer, code string) int { return reportAttempt(writer, code, true) }

func reportAttempt(writer io.Writer, code string, retry bool) int {
	status, message, exit := "refused", "", 1
	switch code {
	case "cancelled":
		status, message, exit = "cancelled", "Cancelled. No files changed.", 0
	case "written":
		status, message, exit = "written", "The reviewed organization plan was created.", 0
	case "usage":
		message, exit = "Use catalog create --id ID --out FILE.", 2
	case "path":
		message = "Choose an absent plan file in an existing safe directory."
	case "data":
		message = "Repair the displayed creation data before saving."
	case "input":
		message = "Use complete valid input within the displayed limits."
	case "output":
		message = "Required output failed before save; no plan was published."
	case "stale":
		message = "The destination context changed; reload and review again."
	case "commit":
		message = "The plan was not published; owned temporary data may need cleanup."
	case "cleanup-refused":
		message = "The plan was not published; temporary cleanup needs repair."
	case "cleanup-written":
		status, message = "written", "The plan was created; temporary cleanup needs repair."
	case "report-written":
		status, message = "written", "The plan was created, but its report could not be written."
	}
	bytes, _ := json.Marshal(struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}{status, "create." + code, message})
	if !writeAll(writer, string(bytes)+"\n") {
		if status == "written" && retry {
			if code == "cleanup-written" {
				reportAttempt(writer, "cleanup-written", false)
			} else {
				reportAttempt(writer, "report-written", false)
			}
		}
		return 1
	}
	return exit
}
