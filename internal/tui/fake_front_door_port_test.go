package tui

import (
	"slices"
	"strings"
	"time"
)

type fakeFrontDoorPort struct {
	plans         []KnownPlan
	listProblem   string
	forgotten     []string
	checks        map[string]FileCheck
	created       []string
	createRefusal string
}

var fakeOpened = time.Date(2026, 11, 3, 16, 45, 0, 0, time.UTC)

func newFakeFrontDoorPort(plans ...KnownPlan) *fakeFrontDoorPort {
	return &fakeFrontDoorPort{plans: plans, checks: map[string]FileCheck{
		"acme.obk.yaml":              {Kind: FilePlan, Path: "/work/acme.obk.yaml"},
		"obk-invitation-sam-lee.zip": {Kind: FileInvitation, Path: "/work/obk-invitation-sam-lee.zip"},
		"notes.txt":                  {Kind: FileRefused, Path: "notes.txt", Refusal: "notes.txt is neither a plan nor an invitation."},
	}}
}

func knownPlans() []KnownPlan {
	return []KnownPlan{
		{Path: "/work/acme/acme.obk.yaml", Name: "acme.obk.yaml", Folder: "/work/acme", Opened: fakeOpened},
		{Path: "/work/globex/platform-team.obk.yaml", Name: "platform-team.obk.yaml", Folder: "/work/globex", Opened: fakeOpened.Add(-49 * time.Hour)},
		{Path: "/old/moved.obk.yaml", Name: "moved.obk.yaml", Folder: "/old", Opened: fakeOpened.Add(-30 * 24 * time.Hour), Missing: true},
	}
}

func (port *fakeFrontDoorPort) KnownPlans() ([]KnownPlan, string) {
	return slices.Clone(port.plans), port.listProblem
}

func (port *fakeFrontDoorPort) Forget(path string) string {
	port.forgotten = append(port.forgotten, path)
	port.plans = slices.DeleteFunc(port.plans, func(plan KnownPlan) bool { return plan.Path == path })
	return ""
}

func (port *fakeFrontDoorPort) Check(path string) FileCheck {
	if checked, found := port.checks[path]; found {
		return checked
	}
	return FileCheck{Kind: FileRefused, Path: path, Refusal: path + " does not exist."}
}

func (port *fakeFrontDoorPort) ProposedPlanPath() string { return "onboarding.obk.yaml" }

func (port *fakeFrontDoorPort) PlanPath(typed string) string {
	if strings.HasPrefix(typed, "/") {
		return typed
	}
	return "/work/" + typed
}

func (port *fakeFrontDoorPort) CreatePlan(path string) PlanCreation {
	if port.createRefusal != "" {
		return PlanCreation{Refusal: port.createRefusal}
	}
	port.created = append(port.created, path)
	return PlanCreation{Path: path}
}

func (port *fakeFrontDoorPort) RecordsFolder() string { return "/work" }

func (port *fakeFrontDoorPort) About() About {
	about := fakeAbout()
	if port.listProblem != "" {
		about.Problems = []string{port.listProblem}
	}
	return about
}

func fakeAbout() About {
	return About{
		Version:     "1.4.0",
		TemplateSet: "1.4.0, the catalog built into this version, which new plans record",
		Places: []AboutPlace{
			{Name: "State home", Path: "/home/sam/.local/state/obk", Chosen: "Chosen as the default for this system. Set OBK_HOME to choose another.", Holds: "Holds the list of your plans, and the sessions and imports of your onboarding.", Use: PlaceReadAndWritten},
			{Name: "Plan list", Path: "/home/sam/.local/state/obk/plans.txt", Holds: "Holds the path of each plan you opened or created here.", Use: PlaceReadAndWritten},
			{Name: "Plan folder", Path: "the folder of each plan you open", Holds: "Holds the plan file.", Use: PlaceReadAndWritten},
			{Name: "Records root", Path: "/work", Holds: "Holds your onboarding record.", Use: PlaceReadAndWritten},
			{Name: "Downloads", Path: "/home/sam/Downloads", Holds: "Later, the tool can look here for an invitation, and only when you choose to.", Use: PlaceNotUsedYet},
		},
	}
}
