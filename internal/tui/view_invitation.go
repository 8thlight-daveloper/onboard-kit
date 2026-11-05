package tui

import (
	"fmt"
	"strings"
)

const (
	handoverLine       = "Give this file to %s. They run: obk %s"
	sendsNothingNotice = "This tool sends nothing. Give the file to the newcomer yourself, through a channel your organization approves."
	untestedNotice     = "Whether the pre-access help can be reached before company access has not been tested."
	invitationNotice   = "An invitation is one file for one newcomer. It holds the plan, their preparation, an empty record, and their first steps guide."
	invitationFixHint  = "Add a person, or fix a preparation, on the Preparation tab. This list checks everyone again each time it opens."
	invitationHelpNote = "Nothing is written until you press y on the last step, and an existing file is never replaced. This tool sends nothing."
	readyText          = "Ready"
)

func (app App) invitationBody(screenFrame frame, colors palette) []string {
	width, _ := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	switch app.invitation.mode {
	case invitationEditing, invitationConfirming:
		return app.invitationFormBody(screenFrame, colors)
	case invitationBlocked:
		return stackPane(pane{lines: clipLines(app.inviteeBlockedLines(width), screenFrame.bodyHeight), focused: true}, width, screenFrame.bodyHeight, colors)
	case invitationWritten:
		return stackPane(pane{lines: clipLines(app.invitationWrittenLines(screenFrame.innerWidth()), screenFrame.bodyHeight), focused: true}, screenFrame.innerWidth(), screenFrame.bodyHeight, colors)
	}
	content, focus := app.inviteeLines(width, colors)
	return app.besideNotices(screenFrame, content, focus, invitationListNotices, colors)
}

func invitationListNotices(width int) []line {
	content := noticeLines(toneBody, invitationNotice, width)
	content = append(content, noticeLines(toneBody, sendsNothingNotice, width)...)
	return append(content, noticeLines(toneFaint, invitationFixHint, width)...)
}

func (person InvitationPerson) readinessText() string {
	switch {
	case person.ready():
		return readyText
	case person.Refusal != InvitationDone:
		return "! not ready, cannot be checked"
	}
	return fmt.Sprintf("! not ready, %d to fix", person.blocking())
}

func (app App) inviteeLines(width int, colors palette) ([]line, int) {
	state := app.invitation
	var content []line
	switch {
	case state.listRefusal != InvitationDone:
		content = append(content, lines(toneProblem, hanging("! problem: ", "The preparations cannot be listed. "+state.listRefusal.sentence(), width))...)
	case len(state.people) == 0:
		content = append(content, lines(toneBody, wrapText("No one has a preparation here yet. Add a person on the Preparation tab, press t.", width))...)
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
		if index == state.cursor {
			focus = len(content)
		}
		text := account + spaces(accountWidth-cellWidth(account)) + summarySeparator + person.readinessText()
		content = append(content, hereLine(text, index == state.cursor, true, width, colors))
	}
	return content, focus
}

func (app App) inviteeBlockedLines(width int) []line {
	person, _ := app.chosenInvitee()
	content := []line{strongLine(person.name() + " is not ready for an invitation"), faintLine("Account " + visibleText(person.Account)), blankLine()}
	if person.Refusal != InvitationDone {
		content = append(content, lines(toneProblem, hanging("! problem: ", person.Refusal.sentence(), width))...)
	}
	content = append(content, reviewGroup("Preparation problems", problemTexts(person.PreparationErrors), toneProblem, width)...)
	content = append(content, reviewGroup("Blockers in the plan", app.blockerTexts(person.Blockers), toneBody, width)...)
	if person.Refusal == InvitationDone && !person.OwnerNamed {
		content = append(content, reviewGroup("Onboarding owner", []string{InvitationOwnerMissing.sentence()}, toneProblem, width)...)
	}
	content = append(content, blankLine())
	return append(content, lines(toneFaint, wrapText("Answer or defer the plan's questions on the Answer tab, and fix the preparation on the Preparation tab. Esc goes back to the list, which checks again.", width))...)
}

func (app App) invitationWrittenLines(width int) []line {
	result := app.invitation.result
	person := app.invitation.form.person
	path := visibleText(result.Path)
	content := []line{faintLine("Invitation written"), blankLine()}
	content = append(content, lines(toneStrong, wrapText(fmt.Sprintf(handoverLine, person.name(), commandPath(path)), width))...)
	content = append(content, blankLine())
	content = append(content, lines(toneBody, wrapText(sendsNothingNotice+" "+untestedNotice, width))...)
	if result.Outcome == InvitationWrittenLeftovers {
		content = append(append(content, blankLine()), lines(toneWarning, hanging("! warning: ", result.Outcome.sentence(), width))...)
	}
	return append(content, blankLine(), faintLine("Enter goes back to the people."))
}

func commandPath(path string) string {
	plain := strings.IndexFunc(path, func(character rune) bool {
		return !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("._-/+@%:,", character))
	}) < 0
	if plain && path != "" {
		return path
	}
	return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
}
