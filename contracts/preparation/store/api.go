package store

import "onboard-kit/internal/model"

type Code string

const (
	CodeListed         Code = "store.listed"
	CodeLoaded         Code = "store.loaded"
	CodeSaved          Code = "store.saved"
	CodeUnchanged      Code = "store.unchanged"
	CodePath           Code = "store.path"
	CodeSource         Code = "store.source"
	CodePlan           Code = "store.plan"
	CodeIdentity       Code = "store.identity"
	CodeLoadedState    Code = "store.loaded-state"
	CodeDraft          Code = "store.draft"
	CodeTooLarge       Code = "store.too-large"
	CodeStale          Code = "store.stale"
	CodeDirectories    Code = "store.directories-missing"
	CodeVolume         Code = "store.volume"
	CodeCommit         Code = "store.commit"
	CodeCleanupRefused Code = "store.cleanup-refused"
	CodeCleanupSaved   Code = "store.cleanup-saved"
)

type Person struct {
	Account     string
	Destination string
	Problem     Code
}

type ListRequest struct {
	Root string
	Plan string
}

type ListResult struct {
	Code   Code
	People []Person
}

type LoadRequest struct {
	Root    string
	Plan    string
	Account string
}

type Loaded interface {
	Account() string
	Destination() string
	Plan() model.Plan
	Preparation() model.Preparation
	Exists() bool
	Bytes() []byte
	DirectoriesMissing() bool
}

type LoadResult struct {
	Code   Code
	Loaded Loaded
}

type SaveRequest struct {
	Loaded            Loaded
	Reviewed          []byte
	CreateDirectories bool
}

type SaveResult struct {
	Code               Code
	Committed          bool
	CreatedDirectories bool
}

type Store interface {
	List(ListRequest) ListResult
	Load(LoadRequest) LoadResult
	Save(SaveRequest) SaveResult
}
