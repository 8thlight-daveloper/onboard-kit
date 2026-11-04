package tui

import (
	"fmt"
	"strings"

	"onboard-kit/internal/model"
)

const (
	notSavedNotice = "Not saved. %s staged. Choose Review and save to write them. Quitting drops them."
	newNotice      = "New. Nothing is written until you review and save."
	foldersNotice  = "Saving creates the folder for this person."
	privacyNotice  = "Anyone who can read this folder can read this file. Keep passwords, secrets, and private details out of it. The terminal also keeps what it shows."
	listNotice     = "A preparation holds what you arrange for one newcomer: their details, and the people and work set up for them. It sits in the newcomer's own folder, beside their record."
	addNotice      = "Press a to add a person. Nothing is written until you review and save."
)

const headingRows = 3

func (app App) preparationBody(screenFrame frame, colors palette) []string {
	switch app.preparation.mode {
	case preparationListing, preparationAdding:
		return app.preparationModeBody(screenFrame, colors)
	}
	width := screenFrame.innerWidth()
	heading := stackPane(pane{lines: append(app.personHeading(width), blankLine()), focused: true}, width, min(headingRows, screenFrame.bodyHeight), colors)
	return append(heading, app.preparationModeBody(app.belowHeading(screenFrame), colors)...)
}

func (app App) belowHeading(screenFrame frame) frame {
	if app.screen != screenPreparation || app.preparation.mode == preparationListing || app.preparation.mode == preparationAdding {
		return screenFrame
	}
	screenFrame.bodyHeight = max(screenFrame.bodyHeight-headingRows, 1)
	return screenFrame
}

func (app App) preparationModeBody(screenFrame frame, colors palette) []string {
	width, _ := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	switch app.preparation.mode {
	case preparationEditing:
		return app.preparationFormBody(screenFrame, colors)
	case preparationReviewing:
		return app.preparationReviewBody(screenFrame, colors)
	case preparationRemoving, preparationDropping:
		return stackPane(pane{lines: clipLines(app.preparationConfirmLines(width), screenFrame.bodyHeight), focused: true}, width, screenFrame.bodyHeight, colors)
	case preparationAdding:
		return stackPane(pane{lines: clipLines(app.addingLines(width, colors), screenFrame.bodyHeight), focused: true}, width, screenFrame.bodyHeight, colors)
	case preparationPerson:
		content, focus := app.personPageLines(width, colors)
		return app.besideNotices(screenFrame, content, focus, app.personNotices, colors)
	}
	content, focus := app.peopleLines(width, colors)
	return app.besideNotices(screenFrame, content, focus, listNotices, colors)
}

func (app App) besideNotices(screenFrame frame, content []line, focus int, notices func(int) []line, colors palette) []string {
	mainWidth, sideWidth := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	height := screenFrame.bodyHeight
	if screenFrame.wide() {
		side := append([]line{faintLine("Notices"), blankLine()}, notices(sideWidth)...)
		return joinPanes(pane{lines: scrollWindow(content, focus, height), focused: true}, pane{lines: clipLines(side, height), focused: true}, mainWidth, sideWidth, height, colors)
	}
	below := notices(mainWidth)
	noticeRows := min(len(below)+2, height/3)
	shown := scrollWindow(content, focus, height-noticeRows)
	combined := append(append(append([]line{}, shown...), blankLine(), faintLine("Notices")), below...)
	return stackPane(pane{lines: clipLines(combined, height), focused: true}, mainWidth, height, colors)
}

func noticeLines(tone tone, text string, width int) []line {
	return append(lines(tone, wrapText(text, width)), blankLine())
}

func listNotices(width int) []line {
	return append(noticeLines(toneBody, listNotice, width), noticeLines(toneFaint, addNotice, width)...)
}

func (app App) personNotices(width int) []line {
	person := app.preparation.person
	var content []line
	for _, id := range person.sharedIDs() {
		content = append(content, lines(toneProblem, hanging("! problem: ", "Two arrangements share the id "+visibleText(id)+". Changes are refused until the file is fixed by hand.", width))...)
		content = append(content, blankLine())
	}
	switch count := len(stagedChanges(person)); {
	case count > 0:
		content = append(content, noticeLines(toneWarning, fmt.Sprintf(notSavedNotice, pluralWords(count, "change is", "changes are")), width)...)
	case !person.loaded.Exists:
		content = append(content, noticeLines(toneWarning, newNotice, width)...)
	}
	if person.loaded.FoldersMissing {
		content = append(content, noticeLines(toneBody, foldersNotice, width)...)
	}
	return append(content, noticeLines(toneFaint, privacyNotice, width)...)
}

func (app App) peopleLines(width int, colors palette) ([]line, int) {
	state := app.preparation
	var content []line
	switch {
	case state.listRefusal != PreparationDone:
		content = append(content, lines(toneProblem, hanging("! problem: ", "The preparations cannot be listed. "+state.listRefusal.sentence(), width))...)
	case len(state.people) == 0:
		content = append(content, bodyLine("No one has a preparation here yet."))
	default:
		content = append(content, strongLine("People with a preparation"))
	}
	content = append(content, blankLine())
	accountWidth := 0
	for _, person := range state.people {
		accountWidth = max(accountWidth, cellWidth(visibleText(person.Account)))
	}
	accountWidth = min(accountWidth, (width-rowMarkWidth)*nameColumnShare/nameColumnOutOf)
	focus := 0
	for index, person := range state.people {
		account := truncateCells(visibleText(person.Account), accountWidth)
		detail := visibleText(person.Destination)
		if person.Unreadable {
			detail = "! cannot be read  " + detail
		}
		if index == state.cursor {
			focus = len(content)
		}
		content = append(content, hereLine(account+spaces(accountWidth-cellWidth(account))+summarySeparator+detail, index == state.cursor, true, width, colors))
	}
	if state.cursor == len(state.people) {
		focus = len(content)
	}
	return append(content, hereLine("Add a person", state.cursor == len(state.people), true, width, colors)), focus
}

func (app App) personHeading(width int) []line {
	loaded := app.preparation.person.loaded
	return []line{strongLine(visibleText(loaded.Account)), faintLine("Saves to " + visibleText(loaded.Destination))}
}

func (app App) personPageLines(width int, colors palette) ([]line, int) {
	person := app.preparation.person
	row := app.preparation.row
	plan := person.loaded.Plan
	var content []line
	focus := 0
	if row == 0 {
		focus = len(content)
	}
	content = append(content, hereLine("Recipient", row == 0, true, width, colors))
	for _, spec := range recipientSpecs() {
		value := displayValues(spec.kind, spec.read(person.draft.Recipient), plan)
		content = append(content, bodyLine(spaces(4)+spec.label+spaces(preparationLabelWidth-cellWidth(spec.label))+value))
	}
	content = append(content, blankLine(), strongLine(fmt.Sprintf("Arrangements (%d)", len(person.arrangements()))))
	idWidth := 0
	for _, arrangement := range person.arrangements() {
		idWidth = max(idWidth, cellWidth(arrangementID(arrangement)))
	}
	idWidth = min(idWidth, (width-rowMarkWidth)*nameColumnShare/nameColumnOutOf)
	for index, arrangement := range person.arrangements() {
		if row == index+1 {
			focus = len(content)
		}
		id := truncateCells(arrangementID(arrangement), idWidth)
		text := id + spaces(idWidth-cellWidth(id)) + summarySeparator + arrangementSummary(arrangement, plan)
		content = append(content, hereLine(text, row == index+1, true, width, colors))
	}
	if row == person.addRow() {
		focus = len(content)
	}
	content = append(content, hereLine("Add an arrangement", row == person.addRow(), true, width, colors), blankLine())
	if row == person.reviewRow() {
		focus = len(content)
	}
	return append(content, hereLine("Review and save", row == person.reviewRow(), true, width, colors)), focus
}

func arrangementID(arrangement model.Arrangement) string {
	if arrangement.ID == "" {
		return "(no id)"
	}
	return visibleText(arrangement.ID)
}

func arrangementSummary(arrangement model.Arrangement, plan model.Plan) string {
	section := displayValues(valueSection, single(arrangement.Section), plan)
	kind := displayValues(valueArrangementKind, single(arrangement.Kind), plan)
	status := displayValues(valueStatus, single(arrangement.Status), plan)
	return strings.Join([]string{section, kind, status}, ", ")
}

func (app App) addingLines(width int, colors palette) []line {
	state := app.preparation
	label := "Account" + spaces(preparationLabelWidth-cellWidth("Account"))
	content := []line{strongLine("Add a person"), blankLine()}
	content = append(content, line{text: colors.render(toneHere, true, "> ") + colors.render(toneStrong, true, label) + state.account.View(), raw: true}, blankLine())
	content = append(content, lines(toneFaint, wrapText("The account names this person's folder, and it never changes. Enter opens the person. Nothing is written until you review and save.", width))...)
	if state.accountProblem != "" {
		content = append(append(content, blankLine()), lines(toneProblem, hanging("! problem: ", state.accountProblem, width))...)
	}
	return content
}

func (app App) preparationConfirmLines(width int) []line {
	person := app.preparation.person
	if app.preparation.mode == preparationDropping {
		var staged []string
		for _, change := range stagedChanges(person) {
			staged = append(staged, change.text)
		}
		content := []line{strongLine("Drop the changes for " + visibleText(person.loaded.Account) + "?"), blankLine()}
		content = append(content, lines(toneBody, wrapText(pluralWords(len(staged), "change is", "changes are")+" staged and not saved. Dropping them leaves the file as it is.", width))...)
		content = append(content, bulletLines("Staged", staged, width)...)
		return append(content, blankLine(), faintLine("Nothing is dropped until you press y."))
	}
	index, _ := person.arrangementRow(app.preparation.row)
	arrangement := person.arrangements()[index]
	content := []line{strongLine("Remove the arrangement " + visibleText(arrangement.ID) + "?"), blankLine()}
	plan := person.loaded.Plan
	facts := "Section: " + displayValues(valueSection, single(arrangement.Section), plan) + ". Kind: " + displayValues(valueArrangementKind, single(arrangement.Kind), plan) + ". Status: " + displayValues(valueStatus, single(arrangement.Status), plan) + "."
	content = append(content, lines(toneBody, wrapText(facts, width))...)
	content = append(content, blankLine())
	content = append(content, lines(toneBody, wrapText("The removal is staged with your other changes. Nothing is written until you review and save.", width))...)
	return append(content, blankLine(), faintLine("Nothing is removed until you press y."))
}
