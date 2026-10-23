package prepare

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	contract "onboard-kit/contracts/command/prepare"
	editorcontract "onboard-kit/contracts/preparation/editor"
	validatorcontract "onboard-kit/contracts/preparation/validator"
	"onboard-kit/internal/model"
	plandecoder "onboard-kit/internal/plan/decoder"
	planvalidator "onboard-kit/internal/plan/validator"
	preparationdecoder "onboard-kit/internal/preparation/decoder"
	preparationeditor "onboard-kit/internal/preparation/editor"
	preparationencoder "onboard-kit/internal/preparation/encoder"
	"onboard-kit/internal/preparation/locator"
	preparationvalidator "onboard-kit/internal/preparation/validator"
	recorddecoder "onboard-kit/internal/record/decoder"
)

const byteLimit = 1 << 20
const usage = "Usage: prepare <init|show|recipient|arrangement|remove> --plan FILE --account ID [--root DIR]\n"

type command struct{}

func New() contract.Command { return command{} }

type arguments struct {
	operation string
	values    map[string]string
	contacts  []string
	help      bool
}

func parse(args []string) (arguments, bool) {
	a := arguments{values: map[string]string{}}
	if len(args) == 0 {
		return a, false
	}
	if args[0] == "--help" {
		a.help = true
		return a, true
	}
	a.operation = args[0]
	allowed := " plan account root "
	switch a.operation {
	case "init", "show":
	case "recipient":
		allowed += "name company-email planned-start-date team manager onboarding-owner buddy work-location time-zone "
	case "arrangement":
		allowed += "id section kind owner contact clear-contacts reference target-date status note "
	case "remove":
		allowed += "id "
	default:
		return a, false
	}
	if len(args) > 1 && args[1] == "--help" {
		a.help = true
		return a, true
	}
	for i := 1; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			return a, false
		}
		name, value, assigned := strings.Cut(strings.TrimPrefix(args[i], "--"), "=")
		if !strings.Contains(allowed, " "+name+" ") {
			return a, false
		}
		if _, exists := a.values[name]; exists && name != "contact" {
			return a, false
		}
		if name == "clear-contacts" {
			if assigned && value != "true" {
				return a, false
			}
			value = "true"
		} else if !assigned {
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "--") {
				return a, false
			}
			value = args[i]
		}
		a.values[name] = value
		if name == "contact" {
			a.contacts = append(a.contacts, value)
		}
	}
	for _, key := range []string{"plan", "account"} {
		if a.values[key] == "" {
			return a, false
		}
	}
	if v, present := a.values["root"]; present && v == "" {
		return a, false
	}
	if _, present := a.values["root"]; !present {
		a.values["root"] = "."
	}
	_, id := a.values["id"]
	_, clear := a.values["clear-contacts"]
	_, contacts := a.values["contact"]
	if clear && contacts {
		return a, false
	}
	changed := len(a.values) - 3
	if a.operation == "recipient" && changed == 0 {
		return a, false
	}
	if a.operation == "arrangement" && (!id || changed < 2) {
		return a, false
	}
	if a.operation == "remove" && !id {
		return a, false
	}
	return a, true
}

func fail(stderr io.Writer, code int, operation, action string) int {
	_, _ = fmt.Fprintf(stderr, "prepare %s: %s\n", operation, action)
	return code
}

func hasErrors(problems []model.Problem) bool {
	for _, p := range problems {
		if p.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func (command) Run(args []string, stdout, stderr io.Writer) int {
	a, valid := parse(args)
	if !valid {
		return fail(stderr, 2, "usage", "check the operation and required flags; use --help")
	}
	if a.help {
		if _, err := io.WriteString(stdout, usage); err != nil {
			return fail(stderr, 1, "help", "could not display help; check output")
		}
		return 0
	}
	reject := func(action string) int { return fail(stderr, 1, a.operation, action) }
	data, err := readPlan(a.values["plan"])
	if err != nil {
		return reject("cannot read plan; select a regular nonsymlink plan within the size limit")
	}
	plan, problems := plandecoder.New().Decode(data)
	if hasErrors(problems) || plan.ID == "" {
		return reject("cannot decode plan identity; repair the plan")
	}
	if hasErrors(planvalidator.New().Validate(plan)) {
		return reject("plan validation failed; repair the plan")
	}
	paths, problems := locator.New().Locate(plan.ResultsPath, a.values["account"])
	if hasErrors(problems) {
		return reject("cannot locate preparation; repair account or plan resultsPath")
	}
	rootName := strings.TrimRight(a.values["root"], string(os.PathSeparator))
	if rootName == "" {
		rootName = string(os.PathSeparator)
	}
	rootInfo, err := os.Lstat(rootName)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return reject("cannot open root; select an existing nonsymlink directory")
	}
	root, err := os.OpenRoot(rootName)
	if err != nil {
		return reject("cannot open root; check directory access")
	}
	defer root.Close()
	openedInfo, err := root.Stat(".")
	if err != nil || !os.SameFile(rootInfo, openedInfo) {
		return reject("root changed while opening; select a stable directory")
	}
	if err = checkPath(root, paths.Preparation); err != nil {
		return reject("unsafe preparation path; repair its directory or file")
	}
	preparation := model.Preparation{SchemaVersion: 1, Kind: "preparation", Account: a.values["account"], PlanID: plan.ID}
	if a.operation == "init" {
		if _, err = root.Lstat(paths.Preparation); !errors.Is(err, os.ErrNotExist) {
			return reject("preparation already exists or cannot be inspected; choose the intended account")
		}
	} else {
		data, err = readArtifact(root, paths.Preparation)
		if err != nil {
			return reject("cannot read preparation; initialize it or repair its file")
		}
		preparation, problems = preparationdecoder.New().Decode(data)
		if hasErrors(problems) || preparation.Account != a.values["account"] || preparation.PlanID != plan.ID {
			return reject("preparation identity or format is invalid; select matching account and plan")
		}
	}
	context := editorcontract.Context{Account: a.values["account"], PlanID: plan.ID}
	for _, c := range plan.Directory.Contacts {
		context.Contacts = append(context.Contacts, c.ID)
	}
	for _, t := range plan.Directory.Teams {
		context.Teams = append(context.Teams, t.ID)
	}
	for _, s := range plan.Sections {
		context.Sections = append(context.Sections, s.ID)
	}
	if a.operation == "show" {
		if err = checkPath(root, paths.Record); err != nil {
			return reject("unsafe record path; repair its directory or file")
		}
		var record *model.Record
		data, err = readArtifact(root, paths.Record)
		if err == nil {
			r, p := recorddecoder.New().Decode(data)
			if hasErrors(p) || r.Account != context.Account || r.Plan.ID != context.PlanID {
				return reject("record identity or format is invalid; select matching account and plan")
			}
			record = &r
		} else if !errors.Is(err, os.ErrNotExist) {
			return reject("cannot read record; repair its file")
		}
		view := showView(preparation, record, context, paths.Preparation, paths.Record)
		if err = json.NewEncoder(stdout).Encode(view); err != nil {
			return reject("could not display preparation; check output")
		}
		return 0
	}
	switch a.operation {
	case "recipient":
		r := preparation.Recipient
		overlayRecipient(&r, a.values)
		preparation, problems = preparationeditor.New().SetRecipient(context, preparation, r)
	case "arrangement":
		item := model.Arrangement{ID: a.values["id"]}
		for _, candidate := range preparation.Arrangements {
			if candidate.ID == item.ID {
				item = candidate
				break
			}
		}
		overlayArrangement(&item, a)
		preparation, problems = preparationeditor.New().SetArrangement(context, preparation, item)
	case "remove":
		preparation, problems = preparationeditor.New().RemoveArrangement(context, preparation, a.values["id"])
	}
	if hasErrors(problems) {
		return reject("preparation change is invalid; repair the supplied fields or duplicate identities")
	}
	data, problems = preparationencoder.New().Encode(preparation)
	if hasErrors(problems) || len(data) > byteLimit {
		return reject("cannot encode preparation within the size limit; repair preparation data")
	}
	if err = publish(root, paths.Preparation, data, a.operation == "init"); err != nil {
		return reject("cannot save preparation; check its destination and permissions")
	}
	if _, err = io.WriteString(stdout, paths.Preparation+"\n"); err != nil {
		return reject("preparation was saved but the result could not be displayed; check output")
	}
	return 0
}

func overlayRecipient(r *model.PreparationRecipient, values map[string]string) {
	fields := map[string]*string{"name": &r.Name, "company-email": &r.CompanyEmail, "team": &r.Team, "manager": &r.Manager, "onboarding-owner": &r.OnboardingOwner, "buddy": &r.Buddy, "work-location": &r.WorkLocation, "time-zone": &r.TimeZone}
	for key, target := range fields {
		if value, present := values[key]; present {
			*target = value
		}
	}
	if value, present := values["planned-start-date"]; present {
		r.PlannedStartDate = model.Date(value)
	}
}

func overlayArrangement(r *model.Arrangement, a arguments) {
	fields := map[string]*string{"section": &r.Section, "kind": &r.Kind, "owner": &r.Owner, "reference": &r.Reference, "status": &r.Status, "note": &r.Note}
	for key, target := range fields {
		if value, present := a.values[key]; present {
			*target = value
		}
	}
	if value, present := a.values["target-date"]; present {
		r.TargetDate = model.Date(value)
	}
	if _, present := a.values["contact"]; present {
		r.Contacts = a.contacts
	}
	if _, present := a.values["clear-contacts"]; present {
		r.Contacts = nil
	}
}

var errArtifact = errors.New("unsafe or oversized artifact")

func readPlan(name string) ([]byte, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errArtifact
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return boundedRead(file)
}

func boundedRead(file *os.File) ([]byte, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errArtifact
	}
	data, err := io.ReadAll(io.LimitReader(file, byteLimit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > byteLimit {
		return nil, errArtifact
	}
	return data, nil
}

func checkPath(root *os.Root, name string) error {
	parts := strings.Split(name, "/")
	for i := range parts {
		info, err := root.Lstat(strings.Join(parts[:i+1], "/"))
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errArtifact
		}
		if i < len(parts)-1 && !info.IsDir() {
			return errArtifact
		}
		if i == len(parts)-1 && !info.Mode().IsRegular() {
			return errArtifact
		}
	}
	return nil
}

func readArtifact(root *os.Root, name string) ([]byte, error) {
	if err := checkPath(root, name); err != nil {
		return nil, err
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return boundedRead(file)
}

func publish(root *os.Root, name string, data []byte, create bool) error {
	directory := path.Dir(name)
	if err := root.MkdirAll(directory, 0700); err != nil {
		return err
	}
	if err := checkPath(root, name); err != nil {
		return err
	}
	parent, err := root.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer parent.Close()
	var random [16]byte
	if _, err = rand.Read(random[:]); err != nil {
		return err
	}
	temporary := ".preparation-" + hex.EncodeToString(random[:])
	file, err := parent.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer parent.Remove(temporary)
	_, writeError := file.Write(data)
	closeError := file.Close()
	if writeError != nil {
		return writeError
	}
	if closeError != nil {
		return closeError
	}
	if create {
		return parent.Link(temporary, path.Base(name))
	}
	return parent.Rename(temporary, path.Base(name))
}

func showView(p model.Preparation, record *model.Record, context editorcontract.Context, preparationPath, recordPath string) map[string]any {
	r := p.Recipient
	recipient := map[string]any{"name": r.Name, "companyEmail": r.CompanyEmail, "plannedStartDate": r.PlannedStartDate, "team": r.Team, "manager": r.Manager, "onboardingOwner": r.OnboardingOwner, "buddy": r.Buddy, "workLocation": r.WorkLocation, "timeZone": r.TimeZone}
	arrangements := make([]map[string]any, 0, len(p.Arrangements))
	for _, a := range p.Arrangements {
		contacts := append([]string{}, a.Contacts...)
		arrangements = append(arrangements, map[string]any{"id": a.ID, "section": a.Section, "kind": a.Kind, "owner": a.Owner, "contacts": contacts, "reference": a.Reference, "targetDate": a.TargetDate, "status": a.Status, "note": a.Note})
	}
	validationContext := validatorcontract.Context{Account: context.Account, PlanID: context.PlanID, Contacts: context.Contacts, Teams: context.Teams, Sections: context.Sections}
	problems := make([]map[string]any, 0)
	for _, problem := range preparationvalidator.New().Validate(validationContext, p) {
		problems = append(problems, map[string]any{"code": problem.Code, "severity": problem.Severity, "message": problem.Message, "path": problem.Detail["path"]})
	}
	outcomes := make([]map[string]any, 0)
	if record != nil {
		for _, o := range record.Outcomes {
			outcomes = append(outcomes, map[string]any{"section": o.Section, "outcome": o.Kind, "on": o.On, "sectionHash": o.SectionHash, "note": o.Note, "differently": o.Differently})
		}
	}
	return map[string]any{"account": p.Account, "planId": p.PlanID, "preparationPath": preparationPath, "recordPath": recordPath, "recipient": recipient, "arrangements": arrangements, "preparationProblems": problems, "record": map[string]any{"present": record != nil, "outcomes": outcomes}}
}
