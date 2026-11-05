package tui

import "time"

type FrontDoorPort interface {
	KnownPlans() ([]KnownPlan, string)
	Forget(path string) string
	Check(path string) FileCheck
	ProposedPlanPath() string
	PlanPath(typed string) string
	MissingFolder(path string) string
	CreatePlan(path string) PlanCreation
	RecordsFolder() string
	About() About
}

type KnownPlan struct {
	Path    string
	Name    string
	Folder  string
	Opened  time.Time
	Missing bool
}

type FileKind int

const (
	FileRefused FileKind = iota
	FilePlan
	FileInvitation
)

type FileCheck struct {
	Kind    FileKind
	Path    string
	Refusal string
}

type PlanCreation struct {
	Path    string
	Refusal string
}

type FrontDoorAction int

const (
	FrontDoorQuit FrontDoorAction = iota
	FrontDoorOpenPlan
	FrontDoorBeginOnboarding
)

type FrontDoorChoice struct {
	Action FrontDoorAction
	Path   string
}
