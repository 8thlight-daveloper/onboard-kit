package places

import (
	"errors"
	"os"
	"path/filepath"

	"onboard-kit/internal/command/registry"
	"onboard-kit/internal/tui"
)

type Context struct {
	Home          Home
	HomeProblem   string
	PlanPath      string
	RecordsRoot   string
	WorkingFolder string
}

type place struct {
	name   string
	path   func(Context) string
	chosen func(Context) string
	holds  func(Context) string
	use    tui.PlaceUse
}

var everyPlace = []place{
	{
		name:   "State home",
		path:   func(context Context) string { return context.Home.Path },
		chosen: homeChoice,
		holds:  fixed("Holds the list of your plans, and the sessions and imports of your onboarding."),
		use:    tui.PlaceReadAndWritten,
	},
	{
		name:  "Plan list",
		path:  inHome(registry.FileName),
		holds: fixed("Holds the path of each plan you opened or created here, with the time it was last opened. Removing a plan from the list never deletes its file."),
		use:   tui.PlaceReadAndWritten,
	},
	{
		name:  "Sessions",
		path:  inHome("sessions"),
		holds: fixed("Holds a copy of the plan and the session of each onboarding you began."),
		use:   tui.PlaceReadAndWritten,
	},
	{
		name:  "Imports",
		path:  inHome("import-pending"),
		holds: fixed("Holds an invitation while it is being opened, so an interrupted start can resume."),
		use:   tui.PlaceReadAndWritten,
	},
	{
		name:  "Plan folder",
		path:  planFolder,
		holds: planFolderHolds,
		use:   tui.PlaceReadAndWritten,
	},
	{
		name:  "Records root",
		path:  recordsRoot,
		holds: recordsRootHolds,
		use:   tui.PlaceReadAndWritten,
	},
	{
		name:  "Downloads",
		path:  downloads,
		holds: fixed("Later, the tool can look here for an invitation, and only when you choose to. Today you type the invitation's path."),
		use:   tui.PlaceNotUsedYet,
	},
}

func List(context Context) []tui.AboutPlace {
	listed := make([]tui.AboutPlace, 0, len(everyPlace))
	for _, each := range everyPlace {
		shown := tui.AboutPlace{Name: each.name, Path: each.path(context), Holds: each.holds(context), Use: each.use}
		if each.chosen != nil {
			shown.Chosen = each.chosen(context)
		}
		listed = append(listed, shown)
	}
	return listed
}

func About(version string, templateSet string, context Context, problems []string) tui.About {
	if context.HomeProblem != "" {
		problems = append([]string{context.HomeProblem}, problems...)
	}
	return tui.About{Version: version, TemplateSet: templateSet, Places: List(context), Problems: problems}
}

func RegistryProblems(home Home) []string {
	if home.Path == "" {
		return nil
	}
	if _, err := registry.Load(home.Path); err != nil {
		return []string{ListProblem(err)}
	}
	return nil
}

func ListProblem(err error) string {
	var damage *registry.Damage
	if errors.As(err, &damage) {
		return "The list of plans at " + damage.File + " cannot be read: " + damage.Reason + ". It is left as it is, and plans you open are not added to it. Remove the file to start a new list."
	}
	return "The list of plans cannot be read: " + err.Error() + "."
}

func fixed(text string) func(Context) string {
	return func(Context) string { return text }
}

func inHome(name string) func(Context) string {
	return func(context Context) string {
		if context.Home.Path == "" {
			return ""
		}
		return filepath.Join(context.Home.Path, name)
	}
}

func homeChoice(context Context) string {
	switch context.Home.Chosen {
	case ChosenByObkHome:
		return "Chosen by the OBK_HOME setting."
	case ChosenByXDGStateHome:
		return "Chosen by the XDG_STATE_HOME setting. Set OBK_HOME to choose another."
	}
	return "Chosen as the default for this system. Set OBK_HOME to choose another."
}

func downloads(Context) string {
	if folder := os.Getenv("XDG_DOWNLOAD_DIR"); folder != "" {
		return folder
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, "Downloads")
	}
	return ""
}

func planFolder(context Context) string {
	if context.PlanPath == "" {
		return "the folder of each plan you open"
	}
	return filepath.Dir(context.PlanPath)
}

func planFolderHolds(context Context) string {
	if context.PlanPath == "" {
		return "Holds the plan file, which the editor saves after each answer, and the invitations the Invitation tab writes."
	}
	return "Holds " + filepath.Base(context.PlanPath) + ", which the editor saves after each answer, and the invitations the Invitation tab writes."
}

func recordsRoot(context Context) string {
	if context.RecordsRoot != "" {
		return context.RecordsRoot
	}
	return context.WorkingFolder
}

func recordsRootHolds(context Context) string {
	if context.RecordsRoot != "" {
		return "Holds a folder for each newcomer of this plan, with their preparation and record."
	}
	return "Holds your onboarding record when you begin your onboarding from this folder. For a plan, the records sit in the plan's folder unless walk --root names another."
}
