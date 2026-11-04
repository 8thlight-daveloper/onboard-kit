package tui

import "onboard-kit/internal/model"

type PreparationPort interface {
	People() ([]PreparationPerson, PreparationOutcome)
	Load(account string) (PreparationLoad, PreparationOutcome)
	Save(account string, reviewed []byte, createFolders bool) PreparationOutcome
}

type PreparationPerson struct {
	Account     string
	Destination string
	Unreadable  bool
}

type PreparationLoad struct {
	Account        string
	Destination    string
	Plan           model.Plan
	Preparation    model.Preparation
	Exists         bool
	Bytes          []byte
	FoldersMissing bool
}

type PreparationOutcome int

const (
	PreparationDone PreparationOutcome = iota
	PreparationUnchanged
	PreparationPathRefused
	PreparationUnreadable
	PreparationPlanWithoutID
	PreparationOtherPerson
	PreparationStale
	PreparationFoldersMissing
	PreparationVolumeNotPrivate
	PreparationNotSaved
	PreparationNotSavedLeftovers
	PreparationSavedLeftovers
)

func (outcome PreparationOutcome) committed() bool {
	return outcome == PreparationDone || outcome == PreparationSavedLeftovers
}

func (outcome PreparationOutcome) sentence() string {
	switch outcome {
	case PreparationUnchanged:
		return "Nothing changed, so nothing was written."
	case PreparationPathRefused:
		return "The folder or the account cannot hold a preparation. Check the account, and that no link or odd file is on the way."
	case PreparationUnreadable:
		return "The plan or this preparation cannot be read. Check the file, or fix it by hand."
	case PreparationPlanWithoutID:
		return "The plan has no id, so no preparation can belong to it."
	case PreparationOtherPerson:
		return "This preparation belongs to another account or another plan."
	case PreparationStale:
		return "The plan or this preparation changed on disk since it was opened. Nothing was saved."
	case PreparationFoldersMissing:
		return "The folder for this person does not exist yet, and it was not created. Nothing was saved."
	case PreparationVolumeNotPrivate:
		return "This disk cannot keep the file private, so nothing was saved. Use a folder on another disk."
	case PreparationNotSaved:
		return "The preparation could not be saved. The files are as they were."
	case PreparationNotSavedLeftovers:
		return "The preparation was not saved, and a temporary file could not be removed from its folder."
	case PreparationSavedLeftovers:
		return "Saved, but a temporary file could not be removed from its folder."
	}
	return ""
}
