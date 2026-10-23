package guide

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	editorcontract "onboard-kit/contracts/preparation/editor"
	contract "onboard-kit/contracts/preparation/guide"
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

const sourceLimit = 1048576

type command struct{}

func New() contract.Command { return command{} }

type options struct{ plan, account, root string }

type snapshot struct {
	data   []byte
	info   os.FileInfo
	absent bool
}

type sourceBodyError struct{}

func (sourceBodyError) Error() string { return "unsupported source body" }

type session struct {
	options           options
	plan              model.Plan
	original          model.Preparation
	draft             model.Preparation
	context           editorcontract.Context
	path              string
	rootInfo          os.FileInfo
	planSource        snapshot
	preparationSource snapshot
	input             io.Reader
	output            io.Writer
}

func (command) Run(args []string, input io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		if write(stdout, "Usage: onboard-kit prepare guide --plan FILE --account ID [--root DIR]\n\nStage recipient details, review the draft, and explicitly save preparation.\nThe plan and onboardee record are not changed. Export is a separate command.\n") != nil {
			return report(stderr, "output")
		}
		return 0
	}
	selected, ok := parse(args)
	if !ok {
		return report(stderr, "usage")
	}
	current := session{options: selected, input: input, output: stdout}
	if outcome := current.load(); outcome != "" {
		return report(stderr, outcome)
	}
	if current.show("Preparation draft — not saved.\nOnboarders maintain preparation; onboardees own their record.\nShared-file readers can read these details. Terminal scrollback can retain them.\nDo not enter secrets or unnecessary personal information.\n") != nil || current.destination() != nil {
		return report(stderr, "output")
	}
	for {
		if current.show("Draft — not saved. Choose recipient, review, or cancel:\n") != nil {
			return report(stderr, "output")
		}
		action, outcome := current.line()
		if outcome != "" {
			return report(stderr, outcome)
		}
		switch action {
		case "cancel":
			return report(stderr, "cancelled")
		case "recipient":
			if outcome := current.form(); outcome != "" {
				return report(stderr, outcome)
			}
		case "review":
			frozen, unchanged, outcome := current.review()
			if outcome != "" {
				return report(stderr, outcome)
			}
			for {
				if current.show("Review complete. Choose save, back, or cancel:\n") != nil {
					return report(stderr, "output")
				}
				action, outcome = current.line()
				if outcome != "" {
					return report(stderr, outcome)
				}
				if action == "cancel" {
					return report(stderr, "cancelled")
				}
				if action == "back" {
					break
				}
				if action == "save" {
					return report(stderr, current.publish(frozen, unchanged))
				}
				if current.invalid() != nil {
					return report(stderr, "output")
				}
			}
		default:
			if current.invalid() != nil {
				return report(stderr, "output")
			}
		}
	}
}

func parse(args []string) (options, bool) {
	result := options{root: "."}
	seen := map[string]bool{}
	for index := 0; index < len(args); index++ {
		name, value, equal := strings.Cut(args[index], "=")
		if name != "--plan" && name != "--account" && name != "--root" || seen[name] {
			return options{}, false
		}
		seen[name] = true
		if !equal {
			index++
			if index >= len(args) || strings.HasPrefix(args[index], "--") {
				return options{}, false
			}
			value = args[index]
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

func safePath(path string) bool {
	if strings.ContainsRune(path, 0) {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return false
		}
	}
	return true
}

func directory(path string) (*os.File, error) {
	if !safePath(path) {
		return nil, unix.EINVAL
	}
	base := "."
	if strings.HasPrefix(path, "/") {
		base = "/"
	}
	fd, err := unix.Open(base, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." {
			continue
		}
		next, openErr := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		closeErr := unix.Close(fd)
		if openErr != nil {
			return nil, openErr
		}
		if closeErr != nil {
			_ = unix.Close(next)
			return nil, closeErr
		}
		fd = next
	}
	return os.NewFile(uintptr(fd), path), nil
}

func parent(root *os.File, path string, create bool) (*os.File, string, error) {
	parts := strings.Split(path, "/")
	fd, err := unix.Dup(int(root.Fd()))
	if err != nil {
		return nil, "", err
	}
	for _, part := range parts[:len(parts)-1] {
		next, openErr := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if errors.Is(openErr, unix.ENOENT) && create {
			if !permissionsSafe(fd, ".") {
				_ = unix.Close(fd)
				return nil, "", unix.EACCES
			}
			if mkdirErr := unix.Mkdirat(fd, part, 0700); mkdirErr != nil && !errors.Is(mkdirErr, unix.EEXIST) {
				_ = unix.Close(fd)
				return nil, "", mkdirErr
			}
			next, openErr = unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		}
		closeErr := unix.Close(fd)
		if openErr != nil {
			return nil, "", openErr
		}
		if closeErr != nil {
			_ = unix.Close(next)
			return nil, "", closeErr
		}
		fd = next
	}
	return os.NewFile(uintptr(fd), "preparation-directory"), parts[len(parts)-1], nil
}

func readAt(root *os.File, path string, allowAbsent bool) (snapshot, error) {
	parentFile, name, err := parent(root, path, false)
	if err != nil {
		if allowAbsent && errors.Is(err, unix.ENOENT) {
			return snapshot{absent: true}, nil
		}
		return snapshot{}, err
	}
	defer parentFile.Close()
	fd, err := unix.Openat(int(parentFile.Fd()), name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		if allowAbsent && errors.Is(err, unix.ENOENT) {
			return snapshot{absent: true}, nil
		}
		return snapshot{}, err
	}
	file := os.NewFile(uintptr(fd), "source")
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return snapshot{}, unix.EINVAL
	}
	if info.Size() > sourceLimit {
		_ = file.Close()
		return snapshot{}, sourceBodyError{}
	}
	data, err := io.ReadAll(io.LimitReader(file, sourceLimit+1))
	closeErr := file.Close()
	if err != nil || closeErr != nil || len(data) > sourceLimit || !utf8.Valid(data) {
		return snapshot{}, sourceBodyError{}
	}
	return snapshot{data: data, info: info}, nil
}

func readPlan(path string) (snapshot, error) {
	if !safePath(path) || strings.HasSuffix(path, "/") {
		return snapshot{}, unix.EINVAL
	}
	position := strings.LastIndex(path, "/")
	parentPath, name := ".", path
	if position >= 0 {
		parentPath, name = path[:position], path[position+1:]
		if parentPath == "" {
			parentPath = "/"
		}
	}
	root, err := directory(parentPath)
	if err != nil {
		return snapshot{}, err
	}
	defer root.Close()
	return readAt(root, name, false)
}

func (current *session) load() string {
	root, err := directory(current.options.root)
	if err != nil {
		return "path"
	}
	defer root.Close()
	current.rootInfo, err = root.Stat()
	if err != nil {
		return "path"
	}
	current.planSource, err = readPlan(current.options.plan)
	if err != nil {
		return "source"
	}
	plan, problems := plandecoder.New().Decode(current.planSource.data)
	if hasErrors(problems) {
		return "source"
	}
	if hasErrors(planvalidator.New().Validate(plan)) || plan.ID == "" {
		return "plan"
	}
	current.plan = plan
	paths, problems := preparationlocator.New().Locate(plan.ResultsPath, current.options.account)
	if len(problems) > 0 {
		return "path"
	}
	current.path = paths.Preparation
	current.preparationSource, err = readAt(root, current.path, true)
	if err != nil {
		var bodyError sourceBodyError
		if errors.As(err, &bodyError) {
			return "source"
		}
		return "path"
	}
	if !current.preparationSource.absent && os.SameFile(current.planSource.info, current.preparationSource.info) {
		return "path"
	}
	preparation := model.Preparation{SchemaVersion: 1, Kind: "preparation", Account: current.options.account, PlanID: plan.ID}
	if !current.preparationSource.absent {
		preparation, problems = preparationdecoder.New().Decode(current.preparationSource.data)
		if hasErrors(problems) {
			return "source"
		}
	}
	context := editorcontract.Context{Account: current.options.account, PlanID: plan.ID}
	for _, contact := range plan.Directory.Contacts {
		context.Contacts = append(context.Contacts, contact.ID)
	}
	for _, team := range plan.Directory.Teams {
		context.Teams = append(context.Teams, team.ID)
	}
	for _, section := range plan.Sections {
		context.Sections = append(context.Sections, section.ID)
	}
	_, problems = preparationeditor.New().SetRecipient(context, preparation, model.PreparationRecipient{})
	if len(problems) > 0 {
		return "identity"
	}
	current.context, current.original, current.draft = context, preparation, preparation
	return ""
}

func hasErrors(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func write(writer io.Writer, value string) error {
	count, err := io.WriteString(writer, value)
	if err != nil {
		return err
	}
	if count != len(value) {
		return io.ErrShortWrite
	}
	return nil
}

func (current *session) show(value string) error { return write(current.output, value) }
func (current *session) invalid() error {
	return current.show("Choose one of the displayed actions.\n")
}
func (current *session) destination() error {
	return current.show("Account: " + quote(current.options.account) + "\nDestination: " + quote(current.path) + "\n")
}

func (current *session) line() (string, string) {
	var result []byte
	var next [1]byte
	emptyReads := 0
	for {
		count, err := current.input.Read(next[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return "", "input"
		}
		if count == 0 {
			if errors.Is(err, io.EOF) {
				return "", "cancelled"
			}
			emptyReads++
			if emptyReads >= 100 {
				return "", "input"
			}
			continue
		}
		emptyReads = 0
		if next[0] != '\n' {
			result = append(result, next[0])
			if len(result) > 65537 {
				return "", "input"
			}
			if errors.Is(err, io.EOF) {
				return "", "cancelled"
			}
			continue
		}
		if len(result) > 0 && result[len(result)-1] == '\r' {
			result = result[:len(result)-1]
		}
		if len(result) > 65536 || !utf8.Valid(result) {
			return "", "input"
		}
		return string(result), ""
	}
}

func labels() []string {
	return []string{"Name", "Company email", "Planned start date (goal)", "Team ID", "Manager contact ID", "Onboarding owner contact ID", "Buddy contact ID", "Work location", "Time zone"}
}
func fields(recipient model.PreparationRecipient) []string {
	return []string{recipient.Name, recipient.CompanyEmail, string(recipient.PlannedStartDate), recipient.Team, recipient.Manager, recipient.OnboardingOwner, recipient.Buddy, recipient.WorkLocation, recipient.TimeZone}
}
func recipient(values []string) model.PreparationRecipient {
	return model.PreparationRecipient{Name: values[0], CompanyEmail: values[1], PlannedStartDate: model.Date(values[2]), Team: values[3], Manager: values[4], OnboardingOwner: values[5], Buddy: values[6], WorkLocation: values[7], TimeZone: values[8]}
}

func (current *session) form() string {
	values := fields(current.draft.Recipient)
	for {
		for index, label := range labels() {
			if current.show(fmt.Sprintf("%d. %s: %s\n", index+1, label, quote(values[index]))) != nil {
				return "output"
			}
		}
		if current.show("Recipient form — not saved. Choose 1-9, apply, or back:\n") != nil {
			return "output"
		}
		action, outcome := current.line()
		if outcome != "" {
			return outcome
		}
		switch action {
		case "back":
			return ""
		case "apply":
			next, problems := preparationeditor.New().SetRecipient(current.context, current.draft, recipient(values))
			if len(problems) == 0 {
				current.draft = next
				return ""
			}
			if current.show("Recipient not applied.\n") != nil || current.problems(problems) != nil {
				return "output"
			}
		default:
			if len(action) == 1 && action[0] >= '1' && action[0] <= '9' {
				if outcome := current.field(values, int(action[0]-'1')); outcome != "" {
					return outcome
				}
			} else if current.invalid() != nil {
				return "output"
			}
		}
	}
}

func (current *session) field(values []string, index int) string {
	for {
		if current.show("Field "+labels()[index]+": "+quote(values[index])+"\nChoose keep, change, clear, or back:\n") != nil {
			return "output"
		}
		action, outcome := current.line()
		if outcome != "" {
			return outcome
		}
		switch action {
		case "keep", "back":
			return ""
		case "clear":
			values[index] = ""
			return ""
		case "change":
			if current.directory(index) != nil {
				return "output"
			}
			for {
				if current.show("Enter the new single-line value (empty input retries):\n") != nil {
					return "output"
				}
				value, outcome := current.line()
				if outcome != "" {
					return outcome
				}
				if value != "" {
					values[index] = value
					return ""
				}
			}
		default:
			if current.invalid() != nil {
				return "output"
			}
		}
	}
}

func (current *session) directory(index int) error {
	if index == 3 {
		if err := current.show("Available teams:\n"); err != nil {
			return err
		}
		if len(current.plan.Directory.Teams) == 0 {
			return current.show("No declared teams.\n")
		}
		for _, team := range current.plan.Directory.Teams {
			if err := current.show("ID: " + quote(team.ID) + " Name: " + quote(team.Name) + "\n"); err != nil {
				return err
			}
		}
	}
	if index >= 4 && index <= 6 {
		if err := current.show("Available contacts:\n"); err != nil {
			return err
		}
		if len(current.plan.Directory.Contacts) == 0 {
			return current.show("No declared contacts.\n")
		}
		for _, contact := range current.plan.Directory.Contacts {
			if err := current.show("ID: " + quote(contact.ID) + " Name: " + quote(contact.Name) + "\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (current *session) problems(problems []model.Problem) error {
	for _, problem := range problems {
		if err := current.show("Severity: " + quote(string(problem.Severity)) + " Code: " + quote(problem.Code) + " Path: " + quote(problem.Detail["path"]) + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func (current *session) problemList(problems []model.Problem) error {
	if len(problems) == 0 {
		return current.show("None.\n")
	}
	return current.problems(problems)
}

func (current *session) review() ([]byte, bool, string) {
	unchanged := !current.preparationSource.absent && reflect.DeepEqual(current.original, current.draft)
	frozen := current.preparationSource.data
	if !unchanged {
		var problems []model.Problem
		frozen, problems = preparationencoder.New().Encode(current.draft)
		if len(problems) > 0 || len(frozen) > sourceLimit {
			return nil, false, "encode"
		}
	}
	review := preparationreviewer.New().Review(reviewercontract.Context{Account: current.options.account}, current.plan, current.draft)
	if current.show("Review — not saved. Preview is not export readiness.\n") != nil || current.destination() != nil {
		return nil, false, "output"
	}
	before, after := fields(current.original.Recipient), fields(current.draft.Recipient)
	for index, label := range labels() {
		line := label + ": " + quote(before[index]) + " -> " + quote(after[index])
		if before[index] != "" && after[index] == "" {
			line += " [cleared]"
		}
		if current.show(line+"\n") != nil {
			return nil, false, "output"
		}
	}
	if current.show("Preparation problems:\n") != nil || current.problemList(review.PreparationProblems) != nil {
		return nil, false, "output"
	}
	if len(review.PreparationProblems) > 0 {
		if current.show("Onboarder blockers: Not evaluated.\nFirst steps: Not evaluated.\n") != nil {
			return nil, false, "output"
		}
		return frozen, unchanged, ""
	}
	if current.show("Onboarder blockers:\n") != nil {
		return nil, false, "output"
	}
	if len(review.OnboarderBlockers) == 0 && current.show("None.\n") != nil {
		return nil, false, "output"
	}
	for _, blocker := range review.OnboarderBlockers {
		if current.show("Question: "+quote(blocker.Question)+"\n") != nil {
			return nil, false, "output"
		}
	}
	if current.show("First-steps problems:\n") != nil || current.problemList(review.FirstStepsProblems) != nil || current.show("First-steps document:\n") != nil {
		return nil, false, "output"
	}
	if review.FirstSteps == "" {
		if current.show("No document returned.\n") != nil {
			return nil, false, "output"
		}
	} else {
		lines := strings.Split(review.FirstSteps, "\n")
		if strings.HasSuffix(review.FirstSteps, "\n") {
			lines = lines[:len(lines)-1]
		}
		for _, line := range lines {
			if current.show("| "+quote(line)+"\n") != nil {
				return nil, false, "output"
			}
		}
	}
	return frozen, unchanged, ""
}

func quote(value string) string {
	var result strings.Builder
	result.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"', '\\':
			result.WriteByte('\\')
			result.WriteRune(character)
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		default:
			if unicode.IsControl(character) || unicode.Is(unicode.Cf, character) || character == '\u2028' || character == '\u2029' {
				if character <= 0xffff {
					fmt.Fprintf(&result, "\\u%04x", character)
				} else {
					first, second := utf16.EncodeRune(character)
					fmt.Fprintf(&result, "\\u%04x\\u%04x", first, second)
				}
			} else {
				result.WriteRune(character)
			}
		}
	}
	result.WriteByte('"')
	return result.String()
}

func sameSnapshot(before, after snapshot) bool {
	if before.absent || after.absent {
		return before.absent == after.absent
	}
	return os.SameFile(before.info, after.info) && bytes.Equal(before.data, after.data) && before.info.Mode() == after.info.Mode()
}

func (current *session) publish(frozen []byte, unchanged bool) string {
	root, err := directory(current.options.root)
	if err != nil {
		return "stale-review"
	}
	defer root.Close()
	info, err := root.Stat()
	if err != nil || !os.SameFile(info, current.rootInfo) {
		return "stale-review"
	}
	planSource, err := readPlan(current.options.plan)
	if err != nil || !sameSnapshot(current.planSource, planSource) {
		return "stale-review"
	}
	preparationSource, err := readAt(root, current.path, true)
	if err != nil || !sameSnapshot(current.preparationSource, preparationSource) {
		return "stale-review"
	}
	if !preparationSource.absent && os.SameFile(planSource.info, preparationSource.info) {
		return "stale-review"
	}
	if unchanged {
		return "unchanged"
	}
	parentFile, name, err := parent(root, current.path, true)
	if err != nil {
		return "commit"
	}
	defer parentFile.Close()
	parentDescriptor := int(parentFile.Fd())
	if !permissionsSafe(parentDescriptor, ".") {
		return "commit"
	}
	if !preparationSource.absent && !permissionsSafe(parentDescriptor, name) {
		return "commit"
	}
	temporary := ""
	var file *os.File
	for index := 0; index < 128; index++ {
		candidate := fmt.Sprintf(".preparation-guide-%d.tmp", index)
		fd, openErr := unix.Openat(parentDescriptor, candidate, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
		if errors.Is(openErr, unix.EEXIST) {
			continue
		}
		if openErr != nil {
			return "commit"
		}
		temporary = candidate
		file = os.NewFile(uintptr(fd), "preparation-temporary")
		break
	}
	if file == nil {
		return "commit"
	}
	cleanup := func() bool {
		err := unix.Unlinkat(parentDescriptor, temporary, 0)
		return err == nil || errors.Is(err, unix.ENOENT)
	}
	var writeErr error
	if !permissionsSafe(parentDescriptor, temporary) {
		writeErr = unix.EACCES
	} else {
		count, err := file.Write(frozen)
		writeErr = err
		if writeErr == nil && count != len(frozen) {
			writeErr = io.ErrShortWrite
		}
	}
	if writeErr == nil && !preparationSource.absent {
		writeErr = file.Chmod(preparationSource.info.Mode().Perm() & 0600)
	}
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		if !cleanup() {
			return "cleanup-refused"
		}
		return "commit"
	}
	if !permissionsSafe(parentDescriptor, temporary) || !permissionsSafe(parentDescriptor, ".") {
		if !cleanup() {
			return "cleanup-refused"
		}
		return "commit"
	}
	if preparationSource.absent {
		err = unix.Linkat(parentDescriptor, temporary, parentDescriptor, name, 0)
	} else {
		err = unix.Renameat(parentDescriptor, temporary, parentDescriptor, name)
	}
	if err != nil {
		if !cleanup() {
			return "cleanup-refused"
		}
		if preparationSource.absent && errors.Is(err, unix.EEXIST) {
			return "stale-review"
		}
		return "commit"
	}
	if !cleanup() {
		return "cleanup-saved"
	}
	return "saved"
}

func report(writer io.Writer, code string) int {
	status, message, exit := "refused", "", 1
	switch code {
	case "cancelled":
		status, message, exit = "cancelled", "Cancelled. No files changed.", 0
	case "saved":
		status, message, exit = "saved", "Preparation saved.", 0
	case "unchanged":
		status, message, exit = "unchanged", "Preparation unchanged. Original bytes preserved.", 0
	case "usage":
		message, exit = "Use prepare guide --plan FILE --account ID [--root DIR].", 2
	case "source":
		message = "Cannot safely load the selected plan or preparation."
	case "plan":
		message = "The selected plan is invalid."
	case "identity":
		message = "Preparation identity or arrangement IDs are invalid."
	case "path":
		message = "Cannot safely select the preparation destination."
	case "input":
		message = "Cannot read a complete supported input line."
	case "output":
		message = "Cannot display the current draft or review. No files changed."
	case "encode":
		message = "Cannot encode the reviewed preparation within the size limit."
	case "stale-review":
		message = "Source files changed. Reload and review before saving."
	case "commit":
		message = "Preparation was not saved. A commit attempt may have left empty preparation directories."
	case "cleanup-refused":
		message = "Preparation was not saved. Owned temporary cleanup needs repair; empty preparation directories may remain."
	case "cleanup-saved":
		status, message = "saved", "Preparation saved. Owned temporary cleanup needs repair."
	case "report-saved":
		status, message = "saved", "Preparation saved, but its completion report could not be written."
	}
	data, _ := json.Marshal(struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}{status, "guide." + code, message})
	if write(writer, string(data)+"\n") != nil {
		if code == "saved" {
			_ = report(writer, "report-saved")
		}
		return 1
	}
	return exit
}
