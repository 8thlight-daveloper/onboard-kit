package tui

import (
	"maps"
	"regexp"
	"slices"
	"testing"

	"onboard-kit/internal/model"
	"onboard-kit/internal/preparation/decoder"
	"onboard-kit/internal/preparation/encoder"
)

type savedPreparation struct {
	account       string
	reviewed      []byte
	createFolders bool
}

type fakePreparations struct {
	plan         model.Plan
	files        map[string][]byte
	saves        []savedPreparation
	saveOutcomes []PreparationOutcome
	loads        int
}

var accountPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func newFakePreparations(plan model.Plan) *fakePreparations {
	return &fakePreparations{plan: plan, files: map[string][]byte{}}
}

func destinationOf(account string) string { return "records/" + account + "/preparation.obk.yaml" }

func (port *fakePreparations) holding(t *testing.T, held model.Preparation) *fakePreparations {
	t.Helper()
	data, problems := encoder.New().Encode(held)
	if len(problems) > 0 {
		t.Fatalf("the held preparation does not encode: %v", problems)
	}
	port.files[held.Account] = data
	return port
}

func (port *fakePreparations) People() ([]PreparationPerson, PreparationOutcome) {
	var people []PreparationPerson
	for _, account := range slices.Sorted(maps.Keys(port.files)) {
		people = append(people, PreparationPerson{Account: account, Destination: destinationOf(account)})
	}
	return people, PreparationDone
}

func (port *fakePreparations) Load(account string) (PreparationLoad, PreparationOutcome) {
	port.loads++
	if !accountPattern.MatchString(account) {
		return PreparationLoad{}, PreparationPathRefused
	}
	loaded := PreparationLoad{Account: account, Destination: destinationOf(account), Plan: port.plan}
	data, exists := port.files[account]
	if !exists {
		loaded.FoldersMissing = true
		loaded.Preparation = model.Preparation{SchemaVersion: 1, Kind: "preparation", Account: account, PlanID: port.plan.ID}
		return loaded, PreparationDone
	}
	held, problems := decoder.New().Decode(data)
	if len(problems) > 0 {
		return PreparationLoad{}, PreparationUnreadable
	}
	loaded.Preparation, loaded.Exists, loaded.Bytes = held, true, slices.Clone(data)
	return loaded, PreparationDone
}

func (port *fakePreparations) Save(account string, reviewed []byte, createFolders bool) PreparationOutcome {
	port.saves = append(port.saves, savedPreparation{account: account, reviewed: slices.Clone(reviewed), createFolders: createFolders})
	if len(port.saveOutcomes) > 0 {
		outcome := port.saveOutcomes[0]
		port.saveOutcomes = port.saveOutcomes[1:]
		return outcome
	}
	if slices.Equal(port.files[account], reviewed) {
		return PreparationUnchanged
	}
	port.files[account] = slices.Clone(reviewed)
	return PreparationDone
}

func (port *fakePreparations) saved(t *testing.T, account string) model.Preparation {
	t.Helper()
	held, problems := decoder.New().Decode(port.files[account])
	if len(problems) > 0 {
		t.Fatalf("the saved bytes for %s do not decode: %v", account, problems)
	}
	return held
}

func startPreparing(t *testing.T, port *fakePreparations) (App, *recorder) {
	t.Helper()
	app, recorded := startWith(t, port.plan, model.AudienceOnboarder)
	return resize(app.WithPreparation(port), 120, 30), recorded
}

func examplePreparation(plan model.Plan) model.Preparation {
	return model.Preparation{
		SchemaVersion: 1, Kind: "preparation", Account: "sam-lee", PlanID: plan.ID,
		Recipient: model.PreparationRecipient{Name: "Sam Lee", Team: "platform", Manager: "sam-okafor", WorkLocation: "remote"},
		Arrangements: []model.Arrangement{
			{ID: "laptop", Section: "access", Kind: "equipment", Owner: "alex-rivera", Status: "requested"},
			{ID: "first-pairing", Section: "clone-and-run-tests", Kind: "pairing", Owner: "alex-rivera", Contacts: []string{"sam-okafor"}, Reference: "A first ticket", Status: "pending"},
		},
	}
}
