package tui

import (
	"fmt"
	"strings"
)

const (
	frontDoorTitle      = "onboard-kit"
	problemMark         = "! problem: "
	openedTimeLayout    = "2006-01-02 15:04"
	missingPlanText     = "missing"
	planStatusWidth     = len("opened ") + len(openedTimeLayout)
	buildDescription    = "Write a new plan file for your team's newcomers, with the core set of sections for a developer's onboarding, then answer its questions."
	openDescription     = "Open a plan file, or an invitation file, by its path."
	beginDescription    = "Open the invitation file your onboarding owner gave you."
	openFileNotice      = "A plan opens in the plan editor. An invitation begins your onboarding."
	beginPathNotice     = "Type the path of the invitation file you were given. Its name starts with obk-invitation and ends with .zip."
	beginTypedNotice    = "After enter, the tool asks you to confirm who you are, then shows your onboarding in this terminal. You type each reply."
	listDamagedNotice   = "The list of your plans cannot be read. Press a to see why."
	newPlanNotice       = "The plan starts with the core set of sections for a developer's onboarding: team and culture, access and setup, learning the code, practice, and feedback. Switch off what you do not need on the Sections tab."
	newPlanStepNotice   = "Enter shows what will be written. Nothing is written until you press y there."
	newPlanConfirmation = "It holds the core set of sections from the catalog built into this tool. An existing file is never replaced. The plan opens in the editor after it is written."
	forgetNotice        = "The file stays where it is. Only this list forgets it."
)

func (door FrontDoor) View() string {
	screenFrame := door.frame()
	if screenFrame.tooSmall() {
		text := fmt.Sprintf("Please make the window larger, at least %d columns by %d rows.", minimumWidth, minimumHeight)
		rows := wrapText(text, max(door.width, 1))
		return strings.Join(rows[:min(len(rows), max(door.height, 1))], "\n")
	}
	colors := newPalette()
	inner := screenFrame.innerWidth()
	rows := []string{door.headerRow(inner, colors), "", ""}
	rows = append(rows, stackPane(pane{lines: door.bodyLines(screenFrame.bodyHeight, colors), focused: true}, inner, screenFrame.bodyHeight, colors)...)
	rows = append(rows, renderCell(door.status, inner, true, colors), renderHints(door.hints(), inner, colors))
	margin := spaces(sideMargin)
	for index, row := range rows {
		rows[index] = strings.TrimRight(margin+row, " ")
	}
	return strings.Join(rows, "\n")
}

func (door FrontDoor) headerRow(width int, colors palette) string {
	left := colors.render(toneStrong, true, frontDoorTitle)
	right := door.screenName()
	if cellWidth(frontDoorTitle)+2+cellWidth(right) > width {
		return left
	}
	return left + spaces(width-cellWidth(frontDoorTitle)-cellWidth(right)) + colors.render(toneFaint, true, right)
}

func (door FrontDoor) screenName() string {
	switch door.screen {
	case doorOpenFile:
		return "Open a file"
	case doorBegin:
		return "Begin my onboarding"
	case doorNewPlan, doorNewConfirm:
		return "New plan"
	case doorAbout:
		return "About"
	}
	return "Start"
}

func (door FrontDoor) bodyLines(height int, colors palette) []line {
	width := door.readingWidth()
	if door.helpOpen {
		return clipLines(door.helpLines(width), height)
	}
	switch door.screen {
	case doorOpenFile:
		return clipLines(door.fieldScreenLines("Open a file", "File", []line{faintLine(openFileNotice)}, width, colors), height)
	case doorBegin:
		return clipLines(door.fieldScreenLines("Begin my onboarding", "Invitation file", door.beginNotices(width), width, colors), height)
	case doorNewPlan:
		return clipLines(door.newPlanLines(width, colors), height)
	case doorNewConfirm:
		return clipLines(door.newConfirmLines(width), height)
	case doorAbout:
		return aboutWindow(aboutLines(door.about, width), door.aboutScroll, height)
	}
	content, focus := door.startLines(width, colors)
	return scrollWindow(content, focus, height)
}

func (door FrontDoor) startLines(width int, colors palette) ([]line, int) {
	heading := "What would you like to do?"
	if len(door.plans) > 0 {
		heading = "Your plans, last used first"
	}
	content := []line{strongLine(heading), blankLine()}
	if door.listProblem != "" {
		content = append(content, lines(toneProblem, hanging(problemMark, listDamagedNotice, width))...)
		content = append(content, blankLine())
	}
	nameWidth := door.planNameWidth(width)
	focus := 0
	for index, row := range door.rows() {
		if row.action == doorActionNew && len(door.plans) > 0 {
			content = append(content, blankLine())
		}
		if index == door.cursor {
			focus = len(content)
		}
		content = append(content, hereLine(row.text(len(door.plans) > 0, nameWidth), index == door.cursor, true, width, colors))
	}
	content = append(content, blankLine())
	if door.forgetting {
		name := visibleText(door.chosenRow().plan.Name)
		content = append(content, lines(toneStrong, wrapText("Remove "+name+" from the list?", width))...)
		content = append(content, lines(toneBody, wrapText(forgetNotice, width))...)
		content = append(content, blankLine(), faintLine("y removes it from the list. Esc, or any other key, keeps it."))
		return content, focus
	}
	content = append(content, lines(toneFaint, wrapText(door.chosenRow().description(), width))...)
	return content, focus
}

func (door FrontDoor) planNameWidth(width int) int {
	longest := 0
	for _, plan := range door.plans {
		longest = max(longest, cellWidth(visibleText(plan.Name)))
	}
	return min(longest, (width-rowMarkWidth)*nameColumnShare/nameColumnOutOf)
}

func (row doorRow) text(withPlans bool, nameWidth int) string {
	switch row.action {
	case doorActionNew:
		if withPlans {
			return "New plan"
		}
		return "Build an onboarding plan"
	case doorActionOpen:
		return "Open a file"
	case doorActionBegin:
		return "Begin my onboarding"
	}
	name := truncateCells(visibleText(row.plan.Name), nameWidth)
	status := missingPlanText
	if !row.plan.Missing {
		status = "opened " + row.plan.Opened.Format(openedTimeLayout)
	}
	return name + spaces(nameWidth-cellWidth(name)) + summarySeparator + status + spaces(planStatusWidth-cellWidth(status)) + summarySeparator + visibleText(row.plan.Folder)
}

func (row doorRow) description() string {
	switch row.action {
	case doorActionNew:
		return buildDescription
	case doorActionOpen:
		return openDescription
	case doorActionBegin:
		return beginDescription
	}
	if row.plan.Missing {
		return "This plan is not at " + visibleText(row.plan.Path) + " any more. It was moved or deleted. Press x to remove it from the list."
	}
	return "Opens " + visibleText(row.plan.Path) + " in the plan editor."
}

func (door FrontDoor) fieldLine(label string, colors palette) line {
	mark := colors.render(toneHere, true, "> ")
	text := mark + colors.render(toneStrong, true, label+spaces(preparationLabelWidth-cellWidth(label)))
	return line{text: text + door.field.View(), raw: true}
}

func (door FrontDoor) fieldScreenLines(heading string, label string, notices []line, width int, colors palette) []line {
	content := []line{strongLine(heading), blankLine(), door.fieldLine(label, colors), blankLine()}
	content = append(content, notices...)
	return append(content, door.problemLines(width)...)
}

func (door FrontDoor) beginNotices(width int) []line {
	content := lines(toneBody, wrapText(beginPathNotice, width))
	content = append(content, lines(toneBody, wrapText("Your onboarding records are kept in "+visibleText(door.port.RecordsFolder())+", the folder you started this tool in.", width))...)
	content = append(content, blankLine())
	return append(content, lines(toneFaint, wrapText(beginTypedNotice, width))...)
}

func (door FrontDoor) problemLines(width int) []line {
	if door.problem == "" {
		return nil
	}
	return append([]line{blankLine()}, lines(toneProblem, hanging(problemMark, door.problem, width))...)
}

func (door FrontDoor) newPlanLines(width int, colors palette) []line {
	content := []line{strongLine("New plan"), blankLine(), door.fieldLine("File", colors)}
	if typed := strings.TrimSpace(door.field.Value()); typed != "" {
		content = append(content, lines(toneFaint, hanging(spaces(formMarkWidth+preparationLabelWidth), "Writes "+visibleText(door.port.PlanPath(typed)), width))...)
		if missing := door.port.MissingFolder(door.port.PlanPath(typed)); missing != "" {
			content = append(content, lines(toneFaint, hanging(spaces(formMarkWidth+preparationLabelWidth), "Also creates the folder "+visibleText(missing), width))...)
		}
	}
	content = append(content, blankLine())
	content = append(content, lines(toneBody, wrapText(newPlanNotice, width))...)
	content = append(content, lines(toneFaint, wrapText(newPlanStepNotice, width))...)
	return append(content, door.problemLines(width)...)
}

func (door FrontDoor) newConfirmLines(width int) []line {
	content := []line{strongLine("Write this new plan?"), blankLine(), bodyLine("File")}
	content = append(content, lines(toneBody, hanging("  ", visibleText(door.newPath), width))...)
	if missing := door.port.MissingFolder(door.newPath); missing != "" {
		content = append(content, blankLine(), bodyLine("Also creates the folder"))
		content = append(content, lines(toneBody, hanging("  ", visibleText(missing), width))...)
	}
	content = append(content, blankLine())
	content = append(content, lines(toneBody, wrapText(newPlanConfirmation, width))...)
	return append(content, blankLine(), faintLine("Nothing is written until you press y."))
}

func (door FrontDoor) hints() []hint {
	if door.helpOpen {
		return []hint{{"a", "about"}, {"any key", "closes help"}}
	}
	switch door.screen {
	case doorOpenFile:
		return []hint{{"enter", "open"}, {"esc", "back"}, {"f1", "help"}}
	case doorBegin:
		return []hint{{"enter", "begin"}, {"esc", "back"}, {"f1", "help"}}
	case doorNewPlan:
		return []hint{{"enter", "next"}, {"esc", door.backName()}, {"f1", "help"}}
	case doorNewConfirm:
		return []hint{{"y", "write"}, {"esc", "change"}, {"?", "help"}}
	case doorAbout:
		return []hint{{"up/down", "scroll"}, {"esc", "back"}, {"?", "help"}}
	}
	if door.forgetting {
		return []hint{{"y", "remove from list"}, {"esc", "keep"}, {"?", "help"}}
	}
	row := door.chosenRow()
	var hints []hint
	if !row.plan.Missing || row.action != doorActionPlan {
		hints = append(hints, hint{"enter", door.enterName()})
	}
	hints = append(hints, hint{"up/down", "move"})
	if row.action == doorActionPlan {
		hints = append(hints, hint{"x", "remove from list"})
	}
	return append(hints, hint{"a", "about"}, hint{"q", "quit"}, hint{"?", "help"})
}

func (door FrontDoor) backName() string {
	if door.newOnly {
		return "quit"
	}
	return "back"
}

func (door FrontDoor) enterName() string {
	if door.chosenRow().action == doorActionPlan {
		return "open"
	}
	return "choose"
}

func (door FrontDoor) helpLines(width int) []line {
	content := []line{strongLine("Keys")}
	for _, entry := range door.helpEntries() {
		content = append(content, lines(toneBody, hanging("", entry.key+"  "+entry.label, width))...)
	}
	content = append(content, blankLine())
	return append(content, lines(toneFaint, wrapText("Press a now to see the version and every place this tool reads or writes. Any other key closes this help.", width))...)
}

func (door FrontDoor) helpEntries() []hint {
	switch door.screen {
	case doorOpenFile, doorBegin:
		return []hint{{"enter", "check the file and open it"}, {"esc", "back to the first screen"}, {"f1", "this help"}}
	case doorNewPlan:
		return []hint{{"enter, ctrl+s", "show what will be written"}, {"esc", door.backName()}, {"f1", "this help"}}
	case doorNewConfirm:
		return []hint{{"y", "write the plan and open it"}, {"esc", "change the file name"}}
	case doorAbout:
		return []hint{{"up, down, k, j", "scroll"}, {"pgup, pgdn", "scroll a page"}, {"esc, q, a", "back"}}
	}
	if door.forgetting {
		return []hint{{"y", "remove the plan from the list. The file is never deleted"}, {"esc, any other key", "keep it in the list"}}
	}
	entries := []hint{{"up, down, k, j", "move"}, {"enter", "open the plan, or choose what to do"}}
	if len(door.plans) > 0 {
		entries = append(entries, hint{"x", "remove the plan from this list, after you confirm. The file is never deleted"})
	}
	return append(entries, hint{"a", "about this tool and the places it uses"}, hint{"q", "quit"}, hint{"ctrl+c", "quit at any time"})
}
