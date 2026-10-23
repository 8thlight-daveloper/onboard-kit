package guide

import (
	"io"
	"reflect"
	"strings"

	plandecodercontract "onboard-kit/contracts/plan/decoder"
	planvalidatorcontract "onboard-kit/contracts/plan/validator"
	decodercontract "onboard-kit/contracts/preparation/decoder"
	editorcontract "onboard-kit/contracts/preparation/editor"
	encodercontract "onboard-kit/contracts/preparation/encoder"
	contract "onboard-kit/contracts/preparation/guide"
	locatorcontract "onboard-kit/contracts/preparation/locator"
	reviewercontract "onboard-kit/contracts/preparation/reviewer"
	"onboard-kit/internal/model"
	plandecoder "onboard-kit/internal/plan/decoder"
	planvalidator "onboard-kit/internal/plan/validator"
	preparationdecoder "onboard-kit/internal/preparation/decoder"
	preparationeditor "onboard-kit/internal/preparation/editor"
	preparationencoder "onboard-kit/internal/preparation/encoder"
	preparationlocator "onboard-kit/internal/preparation/locator"
	preparationreviewer "onboard-kit/internal/preparation/reviewer"
)

const byteLimit = 1048576

const help = "Usage: onboard-kit prepare guide --plan FILE --account ID [--root DIR]\n\nStage recipient details and support arrangements, review, and explicitly save preparation.\nThe plan and onboardee record are not changed. Export is a separate command.\n"

type command struct {
	planDecoder   plandecodercontract.Decoder
	planValidator planvalidatorcontract.Validator
	decoder       decodercontract.Decoder
	encoder       encodercontract.Encoder
	locator       locatorcontract.Locator
	editor        editorcontract.Editor
	reviewer      reviewercontract.Reviewer
}

func New() contract.Command {
	return command{planDecoder: plandecoder.New(), planValidator: planvalidator.New(), decoder: preparationdecoder.New(), encoder: preparationencoder.New(), locator: preparationlocator.New(), editor: preparationeditor.New(), reviewer: preparationreviewer.New()}
}

type options struct{ plan, account, root string }

func arguments(args []string) (options, bool) {
	result := options{root: "."}
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		name, value, equal := strings.Cut(args[i], "=")
		if name != "--plan" && name != "--account" && name != "--root" || seen[name] {
			return options{}, false
		}
		seen[name] = true
		if !equal {
			i++
			if i == len(args) || strings.HasPrefix(args[i], "--") {
				return options{}, false
			}
			value = args[i]
		}
		if value == "" {
			return options{}, false
		}
		switch name {
		case "--plan":
			result.plan = value
		case "--account":
			result.account = value
		case "--root":
			result.root = value
		}
	}
	return result, seen["--plan"] && seen["--account"]
}

type session struct {
	input       lineReader
	output      io.Writer
	problem     string
	cancelled   bool
	plan        model.Plan
	original    model.Preparation
	draft       model.Preparation
	context     editorcontract.Context
	editor      editorcontract.Editor
	encoder     encodercontract.Encoder
	reviewer    reviewercontract.Reviewer
	files       *sources
	account     string
	destination string
}

func (c command) Run(args []string, input io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		if !writeText(stdout, help) {
			return report(stderr, "guide.output")
		}
		return 0
	}
	selected, ok := arguments(args)
	if !ok {
		return report(stderr, "guide.usage")
	}
	files, code := loadSources(selected)
	if code != "" {
		return report(stderr, code)
	}
	defer files.close()
	plan, problems := c.planDecoder.Decode(files.plan.bytes)
	if hasErrors(problems) {
		return report(stderr, "guide.source")
	}
	if plan.ID == "" || hasErrors(c.planValidator.Validate(plan)) {
		return report(stderr, "guide.plan")
	}
	paths, problems := c.locator.Locate(plan.ResultsPath, selected.account)
	if len(problems) != 0 {
		return report(stderr, "guide.path")
	}
	code = files.selectPreparation(paths.Preparation)
	if code != "" {
		return report(stderr, code)
	}
	original := model.Preparation{SchemaVersion: 1, Kind: "preparation", Account: selected.account, PlanID: plan.ID}
	if files.preparation.exists {
		original, problems = c.decoder.Decode(files.preparation.bytes)
		if len(problems) != 0 {
			return report(stderr, "guide.source")
		}
	}
	context := editorcontract.Context{Account: selected.account, PlanID: plan.ID}
	for _, contact := range plan.Directory.Contacts {
		context.Contacts = append(context.Contacts, contact.ID)
	}
	for _, team := range plan.Directory.Teams {
		context.Teams = append(context.Teams, team.ID)
	}
	for _, section := range plan.Sections {
		context.Sections = append(context.Sections, section.ID)
	}
	_, problems = c.editor.SetRecipient(context, original, model.PreparationRecipient{})
	if len(problems) != 0 {
		return report(stderr, "guide.identity")
	}
	s := session{input: lineReader{reader: input}, output: stdout, plan: plan, original: original, draft: original, context: context, editor: c.editor, encoder: c.encoder, reviewer: c.reviewer, files: files, account: selected.account, destination: paths.Preparation}
	s.line("Preparation draft — not saved.")
	s.line("Onboarders maintain preparation; onboardees own their record.")
	s.line("Shared-file readers can read these details. Terminal scrollback can retain them.")
	s.line("Do not enter secrets or unnecessary personal information.")
	s.line("Interaction limit: 1048576 input bytes per run, including line endings. No partial save.")
	s.identity()
	for s.active() {
		action, valid := s.ask("Draft — not saved. Choose recipient, arrangements, review, or cancel:")
		if !valid {
			break
		}
		switch action {
		case "recipient":
			s.recipientForm()
		case "arrangements":
			s.arrangements()
		case "review":
			if result := s.review(); result != "" {
				return report(stderr, result)
			}
		case "cancel":
			s.cancelled = true
		default:
			s.invalid()
		}
	}
	if s.problem != "" {
		return report(stderr, s.problem)
	}
	return report(stderr, "guide.cancelled")
}

func hasErrors(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func (s *session) active() bool { return s.problem == "" && !s.cancelled }

func (s *session) identity() {
	s.line("Account: " + quote(s.account))
	s.line("Destination: " + quote(s.destination))
}

func (s *session) review() string {
	unchanged := s.files.preparation.exists && reflect.DeepEqual(s.original, s.draft)
	bytes := s.files.preparation.bytes
	if !unchanged {
		var problems []model.Problem
		bytes, problems = s.encoder.Encode(s.draft)
		if len(problems) != 0 || len(bytes) > byteLimit {
			s.problem = "guide.encode"
			return ""
		}
	}
	result := s.reviewer.Review(reviewercontract.Context{Account: s.account}, s.plan, s.draft)
	s.line("Review — not saved. Preview is not export readiness.")
	s.identity()
	oldFields, newFields := recipientValues(s.original.Recipient), recipientValues(s.draft.Recipient)
	for i, label := range recipientLabels {
		line := label + ": " + quote(oldFields[i]) + " -> " + quote(newFields[i])
		if oldFields[i] != "" && newFields[i] == "" {
			line += " [cleared]"
		}
		s.line(line)
	}
	s.arrangementReview("Original arrangements:", s.original.Arrangements)
	s.arrangementReview("Draft arrangements:", s.draft.Arrangements)
	s.line("Preparation problems:")
	s.problems(result.PreparationProblems)
	if len(result.PreparationProblems) != 0 {
		s.line("Onboarder blockers: Not evaluated.")
		s.line("First steps: Not evaluated.")
	} else {
		s.line("Onboarder blockers:")
		if len(result.OnboarderBlockers) == 0 {
			s.line("None.")
		}
		for _, blocker := range result.OnboarderBlockers {
			s.line("Question: " + quote(blocker.Question))
		}
		s.line("First-steps problems:")
		s.problems(result.FirstStepsProblems)
		s.line("First-steps document:")
		if result.FirstSteps == "" {
			s.line("No document returned.")
		} else {
			lines := strings.Split(result.FirstSteps, "\n")
			if lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}
			for _, line := range lines {
				s.line("| " + quote(line))
			}
		}
	}
	for s.active() {
		action, ok := s.ask("Review complete. Choose save, back, or cancel:")
		if !ok {
			return ""
		}
		switch action {
		case "back":
			return ""
		case "cancel":
			s.cancelled = true
			return ""
		case "save":
			if !s.files.unchanged() {
				return "guide.stale-review"
			}
			if unchanged {
				return "guide.unchanged"
			}
			return s.files.publish(bytes)
		default:
			s.invalid()
		}
	}
	return ""
}
