package frontdoor

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"onboard-kit/internal/command/create"
	"onboard-kit/internal/command/places"
	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/command/registry"
	"onboard-kit/internal/tui"
)

const (
	proposedPlanName   = "onboarding"
	planFileSuffix     = ".obk.yaml"
	freeNameAttempts   = 1000
	templateSetMeaning = ", the catalog built into this version, which new plans record"
)

type port struct {
	home        places.Home
	homeProblem string
	version     string
	working     string
	identifier  string
	generatedOn string
}

func newPort(version string, identifier string, generatedOn string) *port {
	opened := &port{version: version, identifier: identifier, generatedOn: generatedOn}
	home, err := places.StateHome()
	opened.home = home
	if err != nil {
		opened.homeProblem = "The state home cannot be found: " + err.Error() + "."
	}
	if working, err := os.Getwd(); err == nil {
		opened.working = working
	}
	return opened
}

func (door *port) KnownPlans() ([]tui.KnownPlan, string) {
	if door.home.Path == "" {
		return nil, ""
	}
	entries, err := registry.Load(door.home.Path)
	if err != nil {
		return nil, places.ListProblem(err)
	}
	known := make([]tui.KnownPlan, 0, len(entries))
	for _, entry := range entries {
		_, missing := os.Stat(entry.Path)
		known = append(known, tui.KnownPlan{
			Path:    entry.Path,
			Name:    filepath.Base(entry.Path),
			Folder:  filepath.Dir(entry.Path),
			Opened:  entry.Opened.Local(),
			Missing: missing != nil,
		})
	}
	return known, ""
}

func (door *port) Forget(path string) string {
	if err := registry.Forget(door.home.Path, path); err != nil {
		return places.ListProblem(err)
	}
	return ""
}

func (door *port) Check(path string) tui.FileCheck {
	checked, _ := Check(withHomeFolder(path))
	return checked
}

func withHomeFolder(typed string) string {
	rest, found := strings.CutPrefix(typed, "~/")
	if !found {
		return typed
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return typed
	}
	return filepath.Join(home, rest)
}

func (door *port) ProposedPlanPath() string {
	for attempt := 1; attempt <= freeNameAttempts; attempt++ {
		name := proposedPlanName + planFileSuffix
		if attempt > 1 {
			name = fmt.Sprintf("%s-%d%s", proposedPlanName, attempt, planFileSuffix)
		}
		if _, err := os.Lstat(filepath.Join(door.working, name)); errors.Is(err, os.ErrNotExist) {
			return name
		}
	}
	return ""
}

func (door *port) PlanPath(typed string) string {
	typed = withHomeFolder(typed)
	if filepath.IsAbs(typed) || door.working == "" {
		return typed
	}
	return filepath.Join(door.working, typed)
}

func (door *port) MissingFolder(path string) string {
	folder := filepath.Dir(path)
	if _, err := os.Stat(folder); errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
		return folder
	}
	return ""
}

func (door *port) CreatePlan(path string) tui.PlanCreation {
	data, refusal := corePlanBytes(door.identifier, door.version, door.generatedOn)
	if refusal != "" {
		return tui.PlanCreation{Refusal: refusal}
	}
	if info, err := os.Stat(filepath.Dir(path)); err == nil && !info.IsDir() {
		return tui.PlanCreation{Refusal: filepath.Dir(path) + " is a file, and a plan needs a folder to go in."}
	}
	if missing := door.MissingFolder(path); missing != "" {
		if err := os.MkdirAll(missing, 0o755); err != nil {
			return tui.PlanCreation{Refusal: "The folder " + missing + " cannot be created: " + reason(err) + "."}
		}
	}
	err := create.WriteNew(path, data)
	switch {
	case errors.Is(err, create.ErrExists):
		return tui.PlanCreation{Refusal: "That file already exists, and a new plan never replaces a file. Choose another name."}
	case err != nil:
		return tui.PlanCreation{Refusal: "The plan cannot be written there: " + reason(err) + "."}
	}
	door.remember(path)
	return tui.PlanCreation{Path: path}
}

func (door *port) remember(path string) {
	if door.home.Path != "" {
		registry.Record(door.home.Path, absolute(path), time.Now())
	}
}

func (door *port) RecordsFolder() string { return door.working }

func (door *port) About() tui.About {
	context := places.Context{Home: door.home, HomeProblem: door.homeProblem, WorkingFolder: door.working}
	return places.About(door.version, door.version+templateSetMeaning, context, places.RegistryProblems(door.home))
}

func corePlanBytes(identifier string, version string, generatedOn string) ([]byte, string) {
	if identifier == "" {
		identifier = create.NewIdentifier()
	}
	if generatedOn == "" {
		generatedOn = time.Now().UTC().Format("2006-01-02")
	}
	plan, problems := create.CorePlan(identifier, version, generatedOn)
	if len(problems) > 0 {
		return nil, "The core set cannot be made into a plan: " + problems[0].Message + "."
	}
	data, ok := planfile.Encode(plan, io.Discard)
	if !ok {
		return nil, "The core plan cannot be written as a plan file."
	}
	return data, ""
}

func reason(err error) string {
	var failure *fs.PathError
	switch {
	case errors.Is(err, fs.ErrPermission):
		return "you do not have permission"
	case errors.As(err, &failure):
		return failure.Err.Error()
	}
	return err.Error()
}
