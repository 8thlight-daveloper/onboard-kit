package walk

import (
	contract "onboard-kit/contracts/preparation/store"
	"onboard-kit/internal/preparation/store"
	"onboard-kit/internal/tui"
)

type preparationFiles struct {
	root   string
	plan   string
	store  contract.Store
	loaded map[string]contract.Loaded
}

func newPreparationFiles(root string, plan string) *preparationFiles {
	return &preparationFiles{root: root, plan: plan, store: store.New(), loaded: map[string]contract.Loaded{}}
}

func (files *preparationFiles) People() ([]tui.PreparationPerson, tui.PreparationOutcome) {
	result := files.store.List(contract.ListRequest{Root: files.root, Plan: files.plan})
	if result.Code != contract.CodeListed {
		return nil, outcomeOf(result.Code)
	}
	people := make([]tui.PreparationPerson, 0, len(result.People))
	for _, person := range result.People {
		people = append(people, tui.PreparationPerson{Account: person.Account, Destination: person.Destination, Unreadable: person.Problem != ""})
	}
	return people, tui.PreparationDone
}

func (files *preparationFiles) Load(account string) (tui.PreparationLoad, tui.PreparationOutcome) {
	result := files.store.Load(contract.LoadRequest{Root: files.root, Plan: files.plan, Account: account})
	if result.Code != contract.CodeLoaded {
		return tui.PreparationLoad{}, outcomeOf(result.Code)
	}
	loaded := result.Loaded
	files.loaded[account] = loaded
	return tui.PreparationLoad{
		Account:        loaded.Account(),
		Destination:    loaded.Destination(),
		Plan:           loaded.Plan(),
		Preparation:    loaded.Preparation(),
		Exists:         loaded.Exists(),
		Bytes:          loaded.Bytes(),
		FoldersMissing: loaded.DirectoriesMissing(),
	}, tui.PreparationDone
}

func (files *preparationFiles) Save(account string, reviewed []byte, createFolders bool) tui.PreparationOutcome {
	result := files.store.Save(contract.SaveRequest{Loaded: files.loaded[account], Reviewed: reviewed, CreateDirectories: createFolders})
	return outcomeOf(result.Code)
}

func outcomeOf(code contract.Code) tui.PreparationOutcome {
	switch code {
	case contract.CodeListed, contract.CodeLoaded, contract.CodeSaved:
		return tui.PreparationDone
	case contract.CodeUnchanged:
		return tui.PreparationUnchanged
	case contract.CodePath:
		return tui.PreparationPathRefused
	case contract.CodeSource:
		return tui.PreparationUnreadable
	case contract.CodePlan:
		return tui.PreparationPlanWithoutID
	case contract.CodeIdentity:
		return tui.PreparationOtherPerson
	case contract.CodeStale:
		return tui.PreparationStale
	case contract.CodeDirectories:
		return tui.PreparationFoldersMissing
	case contract.CodeVolume:
		return tui.PreparationVolumeNotPrivate
	case contract.CodeCleanupRefused:
		return tui.PreparationNotSavedLeftovers
	case contract.CodeCleanupSaved:
		return tui.PreparationSavedLeftovers
	}
	return tui.PreparationNotSaved
}
