package tui

import (
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

var ctrlC = tea.KeyMsg{Type: tea.KeyCtrlC}

func pressDoor(door FrontDoor, keys ...tea.Msg) FrontDoor {
	for _, key := range keys {
		next, _ := door.Update(key)
		door = next.(FrontDoor)
	}
	return door
}

func doorWith(port *fakeFrontDoorPort) FrontDoor {
	return pressDoor(NewFrontDoor(port))
}

func doorView(door FrontDoor) string { return plainText(door.View()) }

func mustShow(t *testing.T, door FrontDoor, texts ...string) {
	t.Helper()
	view := doorView(door)
	for _, text := range texts {
		if !strings.Contains(view, text) {
			t.Fatalf("%q is not on the screen:\n%s", text, view)
		}
	}
}

func mustNotShow(t *testing.T, door FrontDoor, text string) {
	t.Helper()
	if view := doorView(door); strings.Contains(view, text) {
		t.Fatalf("%q is on the screen:\n%s", text, view)
	}
}

func TestWithNoKnownPlanTheFirstScreenAsksWhatToDo(t *testing.T) {
	door := doorWith(newFakeFrontDoorPort())
	mustShow(t, door, "What would you like to do?", "> Build an onboarding plan", "  Begin my onboarding", "enter choose")
	mustNotShow(t, door, "Open a file")
	mustNotShow(t, door, "remove from list")
	door = pressDoor(door, esc)
	if door.finished {
		t.Fatal("esc on the first screen quit the tool")
	}
}

func TestWithKnownPlansTheLastUsedIsChosenAndEnterOpensIt(t *testing.T) {
	port := newFakeFrontDoorPort(knownPlans()...)
	door := doorWith(port)
	mustShow(t, door, "Your plans, last used first", "> acme.obk.yaml", "opened 2026-11-03 16:45", "missing", "New plan", "Open a file", "Begin my onboarding", "x remove from list")
	door = pressDoor(door, enter)
	if !door.finished || door.Choice() != (FrontDoorChoice{Action: FrontDoorOpenPlan, Path: "/work/acme/acme.obk.yaml"}) {
		t.Fatalf("got %+v, finished %v", door.Choice(), door.finished)
	}
}

func TestAMissingPlanDoesNotOpenAndCanBeForgotten(t *testing.T) {
	port := newFakeFrontDoorPort(knownPlans()...)
	door := pressDoor(doorWith(port), down, down)
	mustShow(t, door, "This plan is not at /old/moved.obk.yaml any more.", "up/down move  x remove from list")
	mustNotShow(t, door, "enter open")
	door = pressDoor(door, enter)
	if door.finished {
		t.Fatal("a missing plan was opened")
	}
	mustShow(t, door, "Press x to remove it from the list.")
	door = pressDoor(door, typed("x"))
	mustShow(t, door, "Remove moved.obk.yaml from the list?", "The file stays where it is.", "y remove from list  esc keep  ? help")
	door = pressDoor(door, typed("?"))
	mustShow(t, door, "remove the plan from the list. The file is never deleted")
	door = pressDoor(door, esc)
	mustShow(t, door, "Remove moved.obk.yaml from the list?")
	door = pressDoor(door, typed("y"))
	if !slices.Equal(port.forgotten, []string{"/old/moved.obk.yaml"}) {
		t.Fatalf("forgot %v", port.forgotten)
	}
	mustShow(t, door, "Removed moved.obk.yaml from the list. The file was not touched.")
	mustNotShow(t, door, "missing")
}

func TestAnyOtherKeyKeepsThePlanInTheList(t *testing.T) {
	port := newFakeFrontDoorPort(knownPlans()...)
	door := pressDoor(doorWith(port), typed("x"), esc)
	if len(port.forgotten) != 0 || door.finished {
		t.Fatalf("forgot %v, finished %v", port.forgotten, door.finished)
	}
	mustShow(t, door, "> acme.obk.yaml")
	door = pressDoor(door, typed("G"), typed("x"))
	mustShow(t, door, "Only a plan can be removed from the list.")
	mustNotShow(t, door, "x remove from list")
}

func TestANewPlanIsWrittenOnlyAfterTheLastStep(t *testing.T) {
	port := newFakeFrontDoorPort()
	door := pressDoor(doorWith(port), enter)
	mustShow(t, door, "New plan", "onboarding.obk.yaml", "Writes /work/onboarding.obk.yaml", "core set of sections")
	door = pressDoor(door, tea.KeyMsg{Type: tea.KeyCtrlU}, typed("acme.obk.yaml"), enter)
	mustShow(t, door, "Write this new plan?", "/work/acme.obk.yaml", "Nothing is written until you press y.")
	if len(port.created) != 0 {
		t.Fatal("written before y")
	}
	door = pressDoor(door, esc)
	mustShow(t, door, "Writes /work/acme.obk.yaml")
	door = pressDoor(door, enter, typed("y"))
	if !slices.Equal(port.created, []string{"/work/acme.obk.yaml"}) || door.Choice() != (FrontDoorChoice{Action: FrontDoorOpenPlan, Path: "/work/acme.obk.yaml"}) {
		t.Fatalf("created %v, choice %+v", port.created, door.Choice())
	}
}

func TestARefusedNewPlanReturnsToTheNameWithTheReason(t *testing.T) {
	port := newFakeFrontDoorPort()
	port.createRefusal = "That file already exists, and a new plan never replaces a file. Choose another name."
	door := pressDoor(doorWith(port), enter, enter, typed("y"))
	if door.finished {
		t.Fatal("a refused plan was opened")
	}
	mustShow(t, door, "! problem: That file already exists", "Writes /work/onboarding.obk.yaml")
	door = pressDoor(door, tea.KeyMsg{Type: tea.KeyCtrlU}, enter)
	mustShow(t, door, "Type a file name for the plan")
}

func TestTheNewCommandQuitsFromItsOnlyScreen(t *testing.T) {
	door := pressDoor(NewPlanDoor(newFakeFrontDoorPort(), "team.obk.yaml"))
	mustShow(t, door, "Writes /work/team.obk.yaml", "esc quit")
	door = pressDoor(door, esc)
	if !door.finished || door.Choice() != (FrontDoorChoice{}) {
		t.Fatalf("got %+v, finished %v", door.Choice(), door.finished)
	}
}

func TestOpenAFileChecksItsKindBeforeLeaving(t *testing.T) {
	door := pressDoor(doorWith(newFakeFrontDoorPort(knownPlans()...)), typed("G"), up, enter)
	mustShow(t, door, "Open a file", "A plan opens in the plan editor.")
	door = pressDoor(door, typed("notes.txt"), enter)
	mustShow(t, door, "! problem: notes.txt is neither a plan nor an invitation.")
	if door.finished {
		t.Fatal("a refused file was opened")
	}
	door = pressDoor(door, tea.KeyMsg{Type: tea.KeyCtrlU}, typed("acme.obk.yaml"), enter)
	if door.Choice() != (FrontDoorChoice{Action: FrontDoorOpenPlan, Path: "/work/acme.obk.yaml"}) {
		t.Fatalf("got %+v", door.Choice())
	}
}

func TestAnInvitationOpenedAsAFileShowsWhereRecordsGoBeforeItBegins(t *testing.T) {
	door := pressDoor(doorWith(newFakeFrontDoorPort(knownPlans()...)), typed("G"), up, enter, typed("obk-invitation-sam-lee.zip"), enter)
	mustShow(t, door, "Begin my onboarding", "Your onboarding records are kept in /work, the folder you started this tool in.")
	if door.finished {
		t.Fatal("began before the person saw where the records go")
	}
	door = pressDoor(door, enter)
	if door.Choice() != (FrontDoorChoice{Action: FrontDoorBeginOnboarding, Path: "/work/obk-invitation-sam-lee.zip"}) {
		t.Fatalf("got %+v", door.Choice())
	}
}

func TestBeginMyOnboardingTakesOnlyAnInvitation(t *testing.T) {
	door := pressDoor(doorWith(newFakeFrontDoorPort()), down, enter)
	mustShow(t, door, "Invitation file", "~/Downloads/obk-invitation", "confirm who you are")
	door = pressDoor(door, typed("acme.obk.yaml"), enter)
	mustShow(t, door, "This file is a plan, not an invitation.")
	door = pressDoor(door, esc)
	mustShow(t, door, "What would you like to do?")
}

func TestAboutOpensFromTheFirstScreenAndFromHelp(t *testing.T) {
	door := pressDoor(doorWith(newFakeFrontDoorPort()), typed("a"))
	mustShow(t, door, "About this tool", "1.4.0", "State home", "Chosen as the default", "The tool reads and writes it.", "Lines 1 to")
	door = pressDoor(door, typed("G"))
	mustShow(t, door, "Downloads", "/home/sam/Downloads", "Not used yet.")
	door = pressDoor(door, esc)
	mustShow(t, door, "What would you like to do?")
	door = pressDoor(door, typed("?"))
	mustShow(t, door, "Keys", "about this tool")
	door = pressDoor(door, typed("a"))
	mustShow(t, door, "About this tool")
	door = pressDoor(door, enter, typed("?"), typed("z"))
	mustShow(t, door, "What would you like to do?")
}

func TestHelpWhileTypingOpensWithF1AndAboutReturnsToTheField(t *testing.T) {
	door := pressDoor(doorWith(newFakeFrontDoorPort()), down, enter, typed("half"), tea.KeyMsg{Type: tea.KeyF1})
	mustShow(t, door, "check the file and open it")
	door = pressDoor(door, typed("a"), esc)
	mustShow(t, door, "Invitation file", "half")
}

func TestADamagedListIsNamedOnTheFirstScreenAndExplainedInAbout(t *testing.T) {
	port := newFakeFrontDoorPort()
	port.listProblem = "The list of plans at /state/plans.txt cannot be read: line 2 is not a time, a tab, and a full path."
	door := pressDoor(doorWith(port), down)
	mustShow(t, door, "! problem: The list of your plans cannot be read. Press a to see why.", "What would you like to do?", "> Open a file")
	door = pressDoor(door, typed("a"))
	mustShow(t, door, "! problem: The list of plans at /state/plans.txt cannot be read")
}

func TestCtrlCQuitsFromAnyScreenWithNoChoice(t *testing.T) {
	for _, keys := range [][]tea.Msg{{}, {enter}, {down, enter}, {typed("a")}} {
		door := pressDoor(doorWith(newFakeFrontDoorPort()), append(keys, ctrlC)...)
		if !door.finished || door.Choice() != (FrontDoorChoice{}) {
			t.Errorf("%v: got %+v, finished %v", keys, door.Choice(), door.finished)
		}
	}
}

func TestTheEditorHelpReachesAboutWhenItHasOne(t *testing.T) {
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	about := fakeAbout()
	app = app.WithAbout(&about)
	app = press(app, typed("?"))
	if view := plainView(app); !strings.Contains(view, "Press a now to see the version") || !strings.Contains(view, "a about") {
		t.Fatalf("help does not offer About:\n%s", view)
	}
	app = press(app, typed("a"))
	if view := plainView(app); !strings.Contains(view, "About this tool") || !strings.Contains(view, "Plan list") {
		t.Fatalf("About did not open:\n%s", view)
	}
	app = press(app, esc)
	if view := plainView(app); strings.Contains(view, "About this tool") {
		t.Fatalf("About did not close:\n%s", view)
	}
	without, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	without = press(without, typed("?"), typed("a"))
	if view := plainView(without); strings.Contains(view, "About this tool") || strings.Contains(view, "Press a now") {
		t.Fatalf("an editor without About shows it:\n%s", view)
	}
}

func TestAboutScrollsWithinItsLength(t *testing.T) {
	door := pressDoor(doorWith(newFakeFrontDoorPort()), tea.WindowSizeMsg{Width: 80, Height: 14}, typed("a"))
	start := doorView(door)
	door = pressDoor(door, typed("G"))
	end := doorView(door)
	if start == end || !strings.Contains(end, "Downloads") {
		t.Fatalf("did not scroll to the end:\n%s", end)
	}
	door = pressDoor(door, down, down, typed("g"))
	if doorView(door) != start {
		t.Fatalf("did not scroll back:\n%s", doorView(door))
	}
}
