package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type doorScreen int

const (
	doorStart doorScreen = iota
	doorOpenFile
	doorBegin
	doorNewPlan
	doorNewConfirm
	doorAbout
)

type doorAction int

const (
	doorActionPlan doorAction = iota
	doorActionNew
	doorActionOpen
	doorActionBegin
)

type doorRow struct {
	action doorAction
	plan   KnownPlan
}

type FrontDoor struct {
	port        FrontDoorPort
	plans       []KnownPlan
	listProblem string
	screen      doorScreen
	aboutFrom   doorScreen
	cursor      int
	forgetting  bool
	helpOpen    bool
	field       textinput.Model
	problem     string
	newPath     string
	about       About
	aboutScroll int
	newOnly     bool
	choice      FrontDoorChoice
	finished    bool
	status      line
	width       int
	height      int
}

func NewFrontDoor(port FrontDoorPort) FrontDoor {
	door := FrontDoor{port: port, width: startWidth, height: startHeight}
	return door.reloaded()
}

func NewPlanDoor(port FrontDoorPort, path string) FrontDoor {
	door := NewFrontDoor(port)
	door.newOnly = true
	return door.openNewPlan(path)
}

func (door FrontDoor) Choice() FrontDoorChoice { return door.choice }

func (door FrontDoor) Init() tea.Cmd { return nil }

func (door FrontDoor) reloaded() FrontDoor {
	door.plans, door.listProblem = door.port.KnownPlans()
	door.cursor = min(door.cursor, len(door.rows())-1)
	return door
}

func (door FrontDoor) rows() []doorRow {
	if len(door.plans) == 0 && door.listProblem == "" {
		return []doorRow{{action: doorActionNew}, {action: doorActionBegin}}
	}
	rows := make([]doorRow, 0, len(door.plans)+3)
	for _, plan := range door.plans {
		rows = append(rows, doorRow{action: doorActionPlan, plan: plan})
	}
	return append(rows, doorRow{action: doorActionNew}, doorRow{action: doorActionOpen}, doorRow{action: doorActionBegin})
}

func (door FrontDoor) chosenRow() doorRow {
	rows := door.rows()
	return rows[max(min(door.cursor, len(rows)-1), 0)]
}

func (door FrontDoor) typing() bool {
	switch door.screen {
	case doorOpenFile, doorBegin, doorNewPlan:
		return true
	}
	return false
}

func (door FrontDoor) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := message.(type) {
	case tea.WindowSizeMsg:
		door.width, door.height = typed.Width, typed.Height
		door.field.Width = door.fieldWidth()
		return door, nil
	case tea.KeyMsg:
		return door.updateKey(typed)
	}
	return door, nil
}

func (door FrontDoor) updateKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.String() == "ctrl+c" {
		return door.quit(FrontDoorChoice{})
	}
	door.status = line{}
	if door.helpOpen {
		door.helpOpen = false
		if key.String() == "a" {
			return door.openAbout(), nil
		}
		return door, nil
	}
	switch door.screen {
	case doorOpenFile, doorBegin:
		return door.updateFileField(key)
	case doorNewPlan:
		return door.updateNewPlan(key)
	case doorNewConfirm:
		return door.updateNewConfirm(key)
	case doorAbout:
		return door.updateAbout(key), nil
	}
	if door.forgetting {
		return door.updateForgetting(key), nil
	}
	return door.updateStart(key)
}

func (door FrontDoor) quit(choice FrontDoorChoice) (tea.Model, tea.Cmd) {
	door.choice = choice
	door.finished = true
	return door, tea.Quit
}

func (door FrontDoor) updateStart(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	count := len(door.rows())
	switch key.String() {
	case "up", "k":
		door.cursor = max(door.cursor-1, 0)
	case "down", "j":
		door.cursor = min(door.cursor+1, count-1)
	case "home", "g":
		door.cursor = 0
	case "end", "G":
		door.cursor = count - 1
	case "enter":
		return door.choose()
	case "x", "delete":
		return door.askToForget(), nil
	case "a":
		return door.openAbout(), nil
	case "?":
		door.helpOpen = true
	case "q":
		return door.quit(FrontDoorChoice{})
	}
	return door, nil
}

func (door FrontDoor) choose() (tea.Model, tea.Cmd) {
	row := door.chosenRow()
	switch row.action {
	case doorActionNew:
		return door.openNewPlan(door.port.ProposedPlanPath()), nil
	case doorActionOpen:
		return door.openField(doorOpenFile, ""), nil
	case doorActionBegin:
		return door.openField(doorBegin, ""), nil
	}
	if row.plan.Missing {
		door.status = line{text: "This plan is not where it was. Press x to remove it from the list.", tone: toneWarning}
		return door, nil
	}
	return door.quit(FrontDoorChoice{Action: FrontDoorOpenPlan, Path: row.plan.Path})
}

func (door FrontDoor) askToForget() FrontDoor {
	if door.chosenRow().action != doorActionPlan {
		door.status = line{text: "Only a plan can be removed from the list."}
		return door
	}
	door.forgetting = true
	return door
}

func (door FrontDoor) updateForgetting(key tea.KeyMsg) FrontDoor {
	if key.String() == "?" {
		door.helpOpen = true
		return door
	}
	door.forgetting = false
	if key.String() != "y" {
		return door
	}
	plan := door.chosenRow().plan
	if problem := door.port.Forget(plan.Path); problem != "" {
		door.status = line{text: problem, tone: toneProblem}
		return door
	}
	door = door.reloaded()
	if door.status.text == "" {
		door.status = line{text: "Removed " + visibleText(plan.Name) + " from the list. The file was not touched.", tone: toneStrong}
	}
	return door
}

func (door FrontDoor) openField(screen doorScreen, value string) FrontDoor {
	door.screen = screen
	door.problem = ""
	door.field = newTextField(value)
	door.field.Focus()
	door.field.Width = door.fieldWidth()
	if screen == doorBegin {
		door.field.Placeholder = "e.g. ~/Downloads/obk-invitation-your-name.zip"
	} else {
		door.field.Placeholder = "e.g. acme.obk.yaml"
	}
	return door
}

func (door FrontDoor) openNewPlan(path string) FrontDoor {
	door = door.openField(doorNewPlan, path)
	door.field.Placeholder = "e.g. acme.obk.yaml"
	return door
}

func (door FrontDoor) back() (tea.Model, tea.Cmd) {
	if door.newOnly {
		return door.quit(FrontDoorChoice{})
	}
	door.screen = doorStart
	door.problem = ""
	return door, nil
}

func (door FrontDoor) updateFileField(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		return door.back()
	case "f1":
		door.helpOpen = true
		return door, nil
	case "enter":
		return door.checkFile()
	}
	var command tea.Cmd
	door.field, command = door.field.Update(key)
	door.problem = ""
	return door, command
}

func (door FrontDoor) checkFile() (tea.Model, tea.Cmd) {
	typed := strings.TrimSpace(door.field.Value())
	if typed == "" {
		door.problem = "Type the path of a file."
		return door, nil
	}
	checked := door.port.Check(typed)
	switch {
	case checked.Kind == FilePlan && door.screen == doorOpenFile:
		return door.quit(FrontDoorChoice{Action: FrontDoorOpenPlan, Path: checked.Path})
	case checked.Kind == FileInvitation && door.screen == doorOpenFile:
		return door.openField(doorBegin, typed), nil
	case checked.Kind == FileInvitation:
		return door.quit(FrontDoorChoice{Action: FrontDoorBeginOnboarding, Path: checked.Path})
	case checked.Kind == FilePlan:
		door.problem = "This file is a plan, not an invitation. Open it from the first screen with Open a file."
		return door, nil
	}
	door.problem = checked.Refusal
	return door, nil
}

func (door FrontDoor) updateNewPlan(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		return door.back()
	case "f1":
		door.helpOpen = true
		return door, nil
	case "enter", "ctrl+s":
		typed := strings.TrimSpace(door.field.Value())
		if typed == "" {
			door.problem = "Type a file name for the plan, such as acme.obk.yaml."
			return door, nil
		}
		door.newPath = door.port.PlanPath(typed)
		door.screen = doorNewConfirm
		return door, nil
	}
	var command tea.Cmd
	door.field, command = door.field.Update(key)
	door.problem = ""
	return door, command
}

func (door FrontDoor) updateNewConfirm(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "y":
		created := door.port.CreatePlan(door.newPath)
		if created.Refusal != "" {
			door.screen = doorNewPlan
			door.problem = created.Refusal
			return door, nil
		}
		return door.quit(FrontDoorChoice{Action: FrontDoorOpenPlan, Path: created.Path})
	case "esc":
		door.screen = doorNewPlan
	case "?":
		door.helpOpen = true
	}
	return door, nil
}

func (door FrontDoor) openAbout() FrontDoor {
	door.about = door.port.About()
	door.aboutFrom = door.screen
	door.aboutScroll = 0
	door.screen = doorAbout
	return door
}

func (door FrontDoor) updateAbout(key tea.KeyMsg) FrontDoor {
	switch key.String() {
	case "esc", "q", "a", "enter":
		door.screen = door.aboutFrom
		return door
	case "?":
		door.helpOpen = true
		return door
	}
	height := door.frame().bodyHeight
	if offset, moved := scrolledBy(key.String(), door.aboutScroll, height); moved {
		content := aboutLines(door.about, door.readingWidth())
		door.aboutScroll = max(min(offset, aboutScrollLimit(content, height)), 0)
	}
	return door
}

func (door FrontDoor) frame() frame { return newFrame(door.width, door.height) }

func (door FrontDoor) readingWidth() int {
	return min(door.frame().innerWidth(), readingWidth)
}

func (door FrontDoor) fieldWidth() int {
	return max(door.readingWidth()-formMarkWidth-preparationLabelWidth-1, inputMinimum)
}
